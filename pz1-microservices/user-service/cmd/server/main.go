package main

import (
	"log"
	"net/http"

	"example.com/user-service/internal/middleware"
	"example.com/user-service/internal/user"
)

func main() {
	repo := user.NewRepo()
	handler := user.NewHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/users", handler.ListUsers)
	mux.HandleFunc("/users/", handler.GetUserByID)

	addr := ":8081"
	log.Println("user-service started on", addr)

	if err := http.ListenAndServe(addr, middleware.Logging(mux)); err != nil {
		log.Fatal(err)
	}
}
