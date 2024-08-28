package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {

	loadEvnError := godotenv.Load(".env")
	if loadEvnError != nil {
		return
	}

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT environment variable not set")
	}

	router := chi.NewRouter()

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handleReadiness)
	v1Router.Get("/err", handleErr)

	router.Mount("/v1", v1Router)

	log.Printf("Server starting on port %s", portString)
	serverError := srv.ListenAndServe()
	if serverError != nil {
		log.Fatal(serverError)
	}

	fmt.Println("Port:", portString)
}
