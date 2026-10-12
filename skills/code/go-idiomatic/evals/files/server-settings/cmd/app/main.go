package main

import (
	"fmt"
	"net/http"
	"os"

	"example.com/app/internal/server"
)

func main() {
	s := server.NewServer(":8080")
	if err := s.ListenAndServe(http.NotFoundHandler()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
