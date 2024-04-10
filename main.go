package main

import (
	"Term-api/config"
	"Term-api/router"
	"log"
	"net/http"
)

func main() {
	config := config.New()
	r := router.Router(config.String())
	log.Println("Starting server on port 8080: ....")
	log.Fatal(http.ListenAndServe(":8080", r))

}
