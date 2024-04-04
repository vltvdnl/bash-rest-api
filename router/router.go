package router

import (
	middleware "Term-api/middleware/handlers"
	cmdlist "Term-api/storage/cmd-list"
	cmdlog "Term-api/storage/cmd-log"
	"context"
	"log"

	"github.com/gorilla/mux"
)

func Router(psqlconnection string) *mux.Router { // так не делают, но мне пока похуй, потом поправлю (наверное)
	h := middleware.Handler{}
	var err error
	h.LogHand, err = cmdlog.New(psqlconnection)
	if err != nil {
		log.Fatalf("db is out: %v", err)
	}
	h.ListHand, err = cmdlist.New(psqlconnection)
	if err != nil {
		log.Fatalf("db is out: %v", err)
	}
	if err := h.ListHand.Init(context.TODO()); err != nil { // init сделать вместо этого
		log.Fatal("Table not created")
	}
	if err := h.LogHand.Init(context.TODO()); err != nil {
		log.Fatal("table not created")
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/add-command", h.AddToList).Methods("POST")              // add command to list
	r.HandleFunc("/api/run-commands", h.RunCmd).Methods("POST")                // run  commands or queue of commands
	r.HandleFunc("/api/show-commands/log/all", h.ShowLogAll).Methods("GET")    // show all commands in list
	r.HandleFunc("/api/show-commands/log/{id}", h.ShowLogCmd).Methods("GET")   // show spec command by id from log
	r.HandleFunc("/api/show-commands/list/all", h.ShowListAll).Methods("GET")  // show all commands from list
	r.HandleFunc("/api/show-commands/list/{id}", h.ShowListCmd).Methods("GET") // show spec command by id from list
	r.HandleFunc("/api/run-commands/list", h.RunList).Methods("GET")           // run all commands from list
	return r
}
