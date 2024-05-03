package middleware_test

import (
	middleware "Term-api/middleware/handlers"
	"Term-api/storage"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"
)

type mockCMDLog struct {
	CommandsExpected *[]storage.Command
	Err              error
}
type mockExecutor struct{}

func (m *mockExecutor) Start(c *exec.Cmd) error {
	return nil
}
func (m *mockExecutor) Wait(c *exec.Cmd) error {
	return nil
}
func (m *mockCMDLog) ShowAll(ctx context.Context) (*[]storage.Command, error) {
	return m.CommandsExpected, m.Err
}
func (m *mockCMDLog) PickByID(ctx context.Context, id int) (*storage.Command, error) {
	return &(*m.CommandsExpected)[0], m.Err
}
func (m *mockCMDLog) Save(ctx context.Context, c *storage.Command) error {
	return nil
}
func (m *mockCMDLog) Update(ctx context.Context, c *storage.Command) error {
	return nil
}
func (m *mockCMDLog) Close(ctx context.Context) error {
	return nil
}
func TestAllLog(t *testing.T) {
	commandsExpected := []storage.Command{
		{
			ID:         1,
			FullScript: "ls -la",
			CMDStatus:  "success",
			Output:     "mock output",
			Start:      time.Now(),
			End:        time.Now(),
		},
	}
	mockCMDLog := &mockCMDLog{
		CommandsExpected: &commandsExpected,
		Err:              nil,
	}
	handler := middleware.Handler{
		LogHand: mockCMDLog,
		Running: make(chan bool, 1),
	}
	req, err := http.NewRequest("GET", "/api/show-commands", nil)
	if err != nil {
		t.Fatalf("пошла пизда по кочкам %v ", err)
	}
	w := httptest.NewRecorder()
	http.HandlerFunc(handler.AllLog).ServeHTTP(w, req.WithContext(context.Background()))

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	got := make([]storage.Command, 0)
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if len(got) != 1 || got[0].FullScript != commandsExpected[0].FullScript {
		t.Errorf("handler returned unexpected body: got %v want %v", got, commandsExpected)
	}
}
func TestRunCommand(t *testing.T) {
	command := &storage.Command{
		FullScript: `echo "Hello World"`,
	}
	stderrbuff := bytes.Buffer{}
	stdoutbuff := bytes.Buffer{}

	h := middleware.Handler{
		Executor: &mockExecutor{},
		LogHand:  &mockCMDLog{},
		Running:  make(chan bool, 1),
	}
	h.RunCommand(command, stderrbuff, stdoutbuff)
	if command.CMDStatus != storage.Success.String() {
		t.Errorf("Expected command to be successfull")
	}
}
func TestAddCMD(t *testing.T) {
	handler := middleware.Handler{
		Executor: &mockExecutor{},
		LogHand:  &mockCMDLog{},
		Running:  make(chan bool, 1),
	}

	command := storage.Command{
		FullScript: "echo 'test'",
	}
	body, _ := json.Marshal(command)
	req := httptest.NewRequest("POST", "/add-command", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	http.HandlerFunc(handler.AddCMD).ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Result().StatusCode)
	}

	expectedMessage := "Your command is added, it's id: 0"
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if !bytes.Contains(body, []byte(expectedMessage)) {
		t.Errorf("Expected message %v got %v", expectedMessage, string(body))
	}

}

func TestLogCmd(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		want           string
		wantStatusCode int
	}{
		{
			name:           "valid ID",
			path:           "/api/show-commands/1",
			want:           "command details",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid ID",
			path:           "/api/show-commands/notAnInt",
			want:           "Wrong id format",
			wantStatusCode: http.StatusBadRequest,
		},
	}
	h := middleware.Handler{
		Executor: &mockExecutor{},
		LogHand: &mockCMDLog{
			CommandsExpected: &[]storage.Command{
				{
					ID:         1,
					FullScript: "command",
					CMDStatus:  "Success",
					Output:     "command details",
					Start:      time.Now(),
					End:        time.Now(),
				},
			},
			Err: nil,
		},
		Running: make(chan bool, 1),
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			http.HandlerFunc(h.LogCMD).ServeHTTP(w, req)
			if status := w.Code; status != tc.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.wantStatusCode)
			}

			if tc.wantStatusCode == http.StatusBadRequest {
				body, err := io.ReadAll(w.Result().Body)
				if err != nil {
					t.Fatalf("can't scan request body: %v", err)
				}
				if !bytes.Contains(body, []byte(tc.want)) {
					t.Errorf("want %s get %s", string(body), tc.want)
				}
			} else {
				if w.Result().Header.Get("Content-Type") == "application/json" {
					got := storage.Command{}
					if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
						t.Errorf("Unexpected problems with marshaling json: %v", err)
					}
					if got.Output != tc.want {
						t.Errorf("handler returned unexpected body: got %v, want %v", got.Output, tc.want)
					}
				} else {
					body, err := io.ReadAll(w.Result().Body)
					if err != nil {
						t.Fatalf("can't scan request body: %v", err)
					}
					if !bytes.Contains(body, []byte(tc.want)) {
						t.Errorf("want %s get %s", string(body), tc.want)
					}
				}

			}

		})

	}
}
