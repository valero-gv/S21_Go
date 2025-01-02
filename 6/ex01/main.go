package main

import (
	"6/ex01/internal/db"
	"6/ex01/internal/handlers"
	"6/ex01/internal/middleware"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {

	err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	r.PathPrefix("/ex01/css/").Handler(http.StripPrefix("/ex01/css/", http.FileServer(http.Dir("./ex01/css/"))))

	r.HandleFunc("/", handlers.HandleHome)
	r.HandleFunc("/article/{id}", handlers.HandleArticle).Methods("GET")

	r.HandleFunc("/admin", handlers.HandleAdmin)
	r.HandleFunc("/login", handlers.HandleLogin)
	r.HandleFunc("/logout", handlers.HandleLogout)

	r.HandleFunc("/admin/post", handlers.HandlePostArticle).Methods("POST")
	r.Use(middleware.JWTMiddleware)
	r.Use(middleware.RateLimitMiddleware)

	fmt.Println("Server is running on port 8888")
	log.Fatal(http.ListenAndServe(":8888", r))
}
