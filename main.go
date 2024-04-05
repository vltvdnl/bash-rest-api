package main

import (
	"Term-api/router"
	"fmt"
	"log"
	"net/http"
)

const (
	host     = "db"
	port     = 5432
	user     = "postgres"
	password = "123"
	dbname   = "postgres"
)

func main() {
	/* TODO
	1. Routing
	2. Handlers
	3. Storage
	4. Command
	*/
	r := router.Router(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))
	log.Println("Starting server on port 8080: ....")
	log.Fatal(http.ListenAndServe(":8080", r))

}
