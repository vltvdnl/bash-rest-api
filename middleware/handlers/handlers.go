package middleware

import (
	"Term-api/storage"
	cmdlog "Term-api/storage/cmd-log"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// CMDExecutor исполняет команды, которые вводятся в виде строчки
type CMDExecutor interface {
	Start(cmd *exec.Cmd) error
	Wait(cmd *exec.Cmd) error
}
type RealCMDExecutor struct{}

// Start вызывает метод Start exec.Cmd.Start
func (r *RealCMDExecutor) Start(cmd *exec.Cmd) error {
	return cmd.Start()
}

// Wait вызывает метод exec.Cmd.Wait
func (r *RealCMDExecutor) Wait(cmd *exec.Cmd) error {
	return cmd.Wait()
}

type Handler struct {
	Executor CMDExecutor
	LogHand  storage.Storage
	Running  chan bool
}

var (
	stdoutbuff bytes.Buffer
	stderrbuff bytes.Buffer
)

// AddCMD считывает командку из post-запроса, исполняет её в консоли и записывает в бд
func (h *Handler) AddCMD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	var command storage.Command
	err := json.NewDecoder(r.Body).Decode(&command)
	defer r.Body.Close()
	if err != nil || command.FullScript == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("can't read command from request: %v", err)
		if _, err := w.Write([]byte("can't read command from request")); err != nil {
			log.Printf("Can't send message: %v", err)
		}

		return
	}

	command.Start = time.Now()
	err = h.LogHand.Save(r.Context(), &command)
	if err != nil {
		log.Printf("can't save cmd in log: %v", err)
	}

	if _, err := w.Write([]byte(fmt.Sprintf("Your command is added, it's id: %d", command.ID))); err != nil {
		log.Printf("Can't send message: %v", err)
	}

	select {
	case <-h.Running:
		h.RunCommand(&command, stderrbuff, stdoutbuff)
		h.Running <- true
	default:
		var n_stderrbuff, n_stdoutbuff bytes.Buffer
		go h.RunCommand(&command, n_stderrbuff, n_stdoutbuff)
		return
	}
}

// AllLog выдает в виде get-запроса все команды, которые были введены пользователем за эту сессию
func (h *Handler) AllLog(w http.ResponseWriter, r *http.Request) {
	commands, err := h.LogHand.ShowAll(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf8")
		if _, err := w.Write([]byte("No commands in log")); err != nil {
			log.Printf("Can't send message: %v", err)
		}
		return
	} else {
		json.NewEncoder(w).Encode(&commands)
	}
}

// LogCMD выдает подробности о команде, которую ввел пользователь. Выдача команды происходит по ее id
func (h *Handler) LogCMD(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	prefix := "/api/show-commands/"
	if !strings.HasPrefix(path, prefix) {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(path[len(prefix):])
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf8")
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("Wrong id format")); err != nil {
			log.Printf("Can't send message: %v", err)
		}
		return
	}
	command, err := h.LogHand.PickByID(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf8")
		if _, err := w.Write([]byte("No command with this id in log")); err != nil {
			log.Printf("Can't send message: %v", err)
		}
	} else {
		json.NewEncoder(w).Encode(command)
	}
}

// RunCommand принимает строковое представление команды и исполняет ее в консоли, записывая результат в бд
func (h *Handler) RunCommand(command *storage.Command, stderrbuff, stdoutbuff bytes.Buffer) {
	command.Start = time.Now()
	cmd := exec.Command("bash")
	cmd.Stdout, cmd.Stderr = &stdoutbuff, &stderrbuff
	stdin := bytes.NewBufferString(command.FullScript)
	cmd.Stdin = stdin

	if err := h.Executor.Start(cmd); err != nil {
		log.Printf("Command error: %v", err)
		command.CMDStatus = storage.Failed.String() // error in start
	}

	command.CMDStatus = storage.Running.String()
	if err := h.Executor.Wait(cmd); err != nil {
		log.Printf("Command ending error: %v", err)
		command.CMDStatus = storage.Failed.String() // errror while running
	}
	command.End = time.Now()
	command.CMDStatus = storage.Success.String() // success
	command.Output = stdoutbuff.String()

	err := h.LogHand.Update(context.TODO(), command)
	if err != nil {
		log.Printf("can't save cmd in log: %v", err)
	}
	log.Println("Succes, output: ", stdoutbuff.String())
}

func New(psqlconnection string) *Handler {
	loghand, err := cmdlog.New(psqlconnection)
	if err != nil {
		log.Printf("db is out: %v", err)
	}
	if err := loghand.Init(); err != nil {
		log.Printf("table not created: %v", err)
	}
	ch := make(chan bool, 1)
	ch <- true
	return &Handler{Executor: &RealCMDExecutor{},
		LogHand: loghand,
		Running: ch}
}
