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
	"time"

	"github.com/gorilla/mux"
)

/*Короче, здесь получается чето дохуя всего насыпано, что, по идее, можно сделать меньше и лучше, но пока я этим заниматься не стал,
т.к функционал выше красоты)))) да и чтобы блять просто протестить, что хотя бы эт работает */

type Handler struct {
	LogHand *cmdlog.CMDlog
	Running chan bool
}

var (
	stdoutbuff bytes.Buffer
	stderrbuff bytes.Buffer // буфферы для текущей команды
)

func (h *Handler) AddCMD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var command storage.Command
	err := json.NewDecoder(r.Body).Decode(&command)
	if err != nil {
		log.Printf("can't read command from request: %v", err)
		json.NewEncoder(w).Encode("Troubles with reading your command")
		return
	}
	command.Start = time.Now()
	err = h.LogHand.Save(context.TODO(), &command)
	if err != nil {
		log.Printf("can't save cmd in log: %v", err)
	}
	json.NewEncoder(w).Encode(fmt.Sprintf("Your command is added, it's id: %d", command.ID))
	select {
	case <-h.Running:
		h.runCommand(&command, stderrbuff, stdoutbuff)
		h.Running <- true
	default:
		var n_stderrbuff, n_stdoutbuff bytes.Buffer // если предыдущая ещё идёт
		go h.runCommand(&command, n_stderrbuff, n_stdoutbuff)
	}
}
func (h *Handler) AllLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	commands, err := h.LogHand.ShowAll(context.TODO())
	if err != nil {
		json.NewEncoder(w).Encode("No commands in log")
		return
	} else {
		json.NewEncoder(w).Encode(&commands)
	}
}
func (h *Handler) LogCmd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Wrong id format")
		return
	}
	command, err := h.LogHand.PickByID(context.TODO(), id)
	if err != nil {
		json.NewEncoder(w).Encode("List is empty")
	} else {
		json.NewEncoder(w).Encode(command) // maybe some troulbes
	}
}

/*
	 TODO
		1. Надо сделать постоянную запись состояний команды
		2. Возможно пока команда не выполнена хранить её внутри программы
		3. Реализовать поддержку вывода больших команд (как-то выводить промежуточное состояние потока вывода хз)
*/
func (h *Handler) runCommand(command *storage.Command, stderrbuff, stdoutbuff bytes.Buffer) {
	command.Start = time.Now()
	cmd := exec.Command("bash")
	cmd.Stdout, cmd.Stderr = &stdoutbuff, &stderrbuff
	stdin := bytes.NewBufferString(command.FullScript)
	cmd.Stdin = stdin

	if err := cmd.Start(); err != nil {
		log.Println("Command error: %v", err)
		command.CMDStatus = storage.Failed.String() // error in start
	}

	command.CMDStatus = storage.Running.String()
	if err := cmd.Wait(); err != nil {
		log.Println("Command ending error: %v", err)
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
		log.Fatalf("db is out: %v", err)
	}
	if err := loghand.Init(context.TODO()); err != nil {
		log.Fatalf("table not created: %v", err)
	}
	ch := make(chan bool, 1)
	return &Handler{LogHand: loghand, Running: ch}
}
