package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"example.com/order-service/internal/middleware"
	"example.com/order-service/internal/order"
)

const defaultUserServiceURL = "http://localhost:8081"

func main() {
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = defaultUserServiceURL
	}

	repo := order.NewRepo()
	client := order.NewUserServiceClient(userServiceURL)
	handler := order.NewHandler(repo, client)

	mux := http.NewServeMux()
	mux.HandleFunc("/orders/by-user/", handler.GetOrdersByUserID)
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/full") {
			handler.GetOrderWithUser(w, r)
			return
		}
		handler.GetOrderByID(w, r)
	})

	addr := ":8082"
	log.Println("order-service started on", addr)
	log.Println("user-service URL:", userServiceURL)

	if err := http.ListenAndServe(addr, middleware.Logging(mux)); err != nil {
		log.Fatal(err)
	}
}
