package main

import (
	"log"
	"net/http"
	"os"

	"github.com/NerdAlert58/sudoku-flow2/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, httpapi.New()))
}
