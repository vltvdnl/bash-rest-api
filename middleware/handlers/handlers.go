package middleware

import (
	"Term-api/storage"
	cmdlist "Term-api/storage/cmd-list"
	cmdlog "Term-api/storage/cmd-log"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

/*Короче, здесь получается чето дохуя всего насыпано, что, по идее, можно сделать меньше и лучше, но пока я этим заниматься не стал,
т.к функционал выше красоты)))) да и чтобы блять просто протестить, что хотя бы эт работает */

type Handler struct {
	LogHand  *cmdlog.CMDlog
	ListHand *cmdlist.List
}

func (h *Handler) RunCmd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var command storage.Command
	err := json.NewDecoder(r.Body).Decode(&command)
	if err != nil {
		log.Printf("can't read command from request: %v", err)
		json.NewEncoder(w).Encode("Troubles with reading your command")
		return
	}
	runCommand(w, r, &command)
	if err := h.LogHand.Save(context.TODO(), &command); err != nil {
		log.Printf("can't put command to log: %v", err)
	}
}
func (h *Handler) AddToList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var command storage.Command
	err := json.NewDecoder(r.Body).Decode(&command)
	if err != nil {
		log.Printf("can't read command from request: %v", err)
		json.NewEncoder(w).Encode("Troubles with reading your command")
		return
	}
	command.CreatedAt = time.Now()
	_, err = command.ScriptToCmd()
	if err != nil {
		log.Println("Not valid command: %v", err)
		json.NewEncoder(w).Encode(fmt.Sprintf("Not valid command: %s", command.FullScript))
		return
	} else {
		if err := h.ListHand.Save(context.TODO(), &command); err != nil {
			log.Printf("can't put command to list: %v", err)
			w.WriteHeader(400)
			json.NewEncoder(w).Encode("Problems with list: internal")
		} else {
			json.NewEncoder(w).Encode("Your command saved to list")
		}
	}
}
func (h *Handler) ShowLogAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	commands, err := h.LogHand.ShowALL(context.TODO())
	if err != nil {
		json.NewEncoder(w).Encode("No commands in log")
		return
	} else {
		json.NewEncoder(w).Encode(&commands)
	}
}
func (h *Handler) ShowLogCmd(w http.ResponseWriter, r *http.Request) {
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
func (h *Handler) ShowListAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	commands, err := h.ListHand.ShowALL(context.TODO())
	if err != nil {
		json.NewEncoder(w).Encode("No commands in log")
		return
	} else {
		json.NewEncoder(w).Encode(&commands)
	}
}
func (h *Handler) ShowListCmd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Wrong id format")
		return
	}
	command, err := h.ListHand.PickByID(context.TODO(), id)
	if err != nil {
		log.Println("Error: %v", err)
		json.NewEncoder(w).Encode("List is empty")
	} else {
		json.NewEncoder(w).Encode(command) // maybe some troulbes
	}
}
func (h *Handler) RunList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	commands, err := h.ListHand.ShowALL(context.TODO())
	if err != nil {
		log.Println("Errors: %v", err)
		json.NewEncoder(w).Encode("No commands in log")
		return
	}
	for _, command := range *commands {
		if command.IsParallel {
			go func() {
				runCommand(w, r, &command)
				if err := h.LogHand.Save(context.TODO(), &command); err != nil {
					log.Printf("can't put command to log: %v", err)

				}
			}()
			time.Sleep(1 * time.Millisecond)
			continue
		}
		runCommand(w, r, &command)
		if err := h.LogHand.Save(context.TODO(), &command); err != nil {
			log.Printf("can't put command to log: %v", err)
		}
	}
}

func runCommand(w http.ResponseWriter, r *http.Request, command *storage.Command) {
	command.CreatedAt = time.Now()
	cmd, err := command.ScriptToCmd()
	if err != nil {
		json.NewEncoder(w).Encode("Not valid command")
	} else {
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("can't do your command: %v", err)
			command.Success = false
			json.NewEncoder(w).Encode("Some problems with your command")
		} else {
			command.Success = true
			json.NewEncoder(w).Encode(fmt.Sprintf("Request done, command output: %s", string(output)))
		}
		command.Output = string(output)
	}
}
