package router

import (
	middleware "Term-api/middleware/handlers"

	"github.com/gorilla/mux"
)

func Router(psqlconnection string) *mux.Router { // так не делают, но мне пока похуй, потом поправлю (наверное)
	h := middleware.New(psqlconnection)
	r := mux.NewRouter()
	r.HandleFunc("/api/add-command", h.AddCMD).Methods("POST") // add command to list
	// r.HandleFunc("/api/run-commands", h.RunCmd).Methods("POST")                // run  commands or queue of commands
	r.HandleFunc("/api/show-commands/all", h.AllLog).Methods("GET")  // show all commands in log
	r.HandleFunc("/api/show-commands/{id}", h.LogCmd).Methods("GET") // show spec command by id from log
	// r.HandleFunc("/api/show-commands/list/all", h.ShowListAll).Methods("GET")  // show all commands from list
	// r.HandleFunc("/api/show-commands/list/{id}", h.ShowListCmd).Methods("GET") // show spec command by id from list
	// r.HandleFunc("/api/run-commands/list", h.RunList).Methods("GET")           // run all commands from list
	return r
}
