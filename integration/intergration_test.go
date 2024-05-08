package integration_test

import (
	middleware "Term-api/middleware/handlers"
	"Term-api/router"
	"Term-api/storage"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
)

type mockCMDLog struct {
	Commands []storage.Command
}
type mockExecutor struct{}

func (m *mockExecutor) Start(c *exec.Cmd) error {
	return nil
}
func (m *mockExecutor) Wait(c *exec.Cmd) error {
	return nil
}
func (m *mockCMDLog) ShowAll(ctx context.Context) (*[]storage.Command, error) {
	return nil, fmt.Errorf("No commands in log")
}
func (m *mockCMDLog) PickByID(ctx context.Context, id int) (*storage.Command, error) {
	return nil, fmt.Errorf("can't pick command with this id, maybe id is wrong")
}
func (m *mockCMDLog) Save(ctx context.Context, c *storage.Command) error {
	m.Commands = append(m.Commands, *c)
	c.ID = len(m.Commands)
	return nil
}
func (m *mockCMDLog) Update(ctx context.Context, c *storage.Command) error {
	return nil
}
func (m *mockCMDLog) Close(ctx context.Context) error {
	return nil
}

func TestRouter(t *testing.T) {
	h := middleware.Handler{
		LogHand:  &mockCMDLog{},
		Executor: &mockExecutor{},
		Running:  make(chan bool, 1),
	}
	r := router.Router(&h)
	server := httptest.NewServer(r)
	defer server.Close()

	tests := []struct {
		name               string
		url                string
		query              string
		method             string
		expectedStatusCode int
		expectedMessage    string
	}{
		{
			name:               "Add command: valid command",
			url:                "/api/add-command",
			query:              `{"script": "mkdir aa; cd aa; mkdir jj; mkdir pp; ls"}`,
			method:             http.MethodPost,
			expectedStatusCode: http.StatusOK,
			expectedMessage:    "Your command is added, it's id: 1",
		},
		{
			name:               "Add command: invalid input",
			url:                "/api/add-command",
			query:              `{"scrt": "not valid command"}`,
			method:             http.MethodPost,
			expectedStatusCode: http.StatusBadRequest,
			expectedMessage:    "can't read command from request",
		},
		{
			name:               "Show all commands: valid method",
			url:                "/api/show-commands",
			query:              "",
			method:             http.MethodGet,
			expectedStatusCode: http.StatusOK,
			expectedMessage:    "No commands in log",
		},
		{
			name:               "Show all commands: invalid method",
			url:                "/api/show-commands",
			query:              "",
			method:             http.MethodPost,
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedMessage:    "Method is not allowed",
		},
		{
			name:               "Show command by id: valid input",
			url:                "/api/show-commands/1",
			query:              "",
			method:             http.MethodGet,
			expectedStatusCode: http.StatusOK,
			expectedMessage:    "",
		},
		{
			name:               "Show command by id: invalid input",
			url:                "/api/show-commands/notValidId",
			query:              "",
			expectedStatusCode: http.StatusBadRequest,
			expectedMessage:    "Wrong id format",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var req *http.Request
			var err error
			if test.query != "" {
				req, err = http.NewRequest(test.method, server.URL+test.url, bytes.NewBuffer([]byte(test.query)))
			} else {
				req, err = http.NewRequest(test.method, server.URL+test.url, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			if req.Body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != test.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", test.expectedStatusCode, resp.StatusCode)
			}
			defer resp.Body.Close()

			if resp.Body == nil {
				t.Fatalf("response body is nil")
			}

			if resp.Header.Get("Content-Type") == "application/json" {
				var gotMessage []storage.Command
				if err := json.NewDecoder(resp.Body).Decode(&gotMessage); err != nil {
					t.Errorf("Unexpected problems with marshaling json: %v", err)
				}
				if len(gotMessage) == 0 || gotMessage[0].FullScript != test.expectedMessage {
					t.Errorf("JSON: Expected message %v got %v", test.expectedMessage, gotMessage)
				}
			} else {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				if !bytes.Contains(body, []byte(test.expectedMessage)) {
					t.Errorf("Expected message %v got %v", test.expectedMessage, string(body))
				}

			}
		})
	}
}
