package main

import (
	"finalProject/database"
	"finalProject/handlers"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const webDir = "./web"
const port = 7540

func main() {
	database.CreateDB()
	r := chi.NewRouter()
	fs := http.FileServer(http.Dir(webDir))

	r.Handle("/*", fs)
	log.Printf("Loaded frontend from %s\n", webDir)

	r.Get("/api/nextdate", handlers.GetNextDate)
	r.Post("/api/task", handlers.PostTaskHandler)
	r.Get("/api/tasks", handlers.GetTasksHandler)
	r.Get("/api/task", handlers.GetTaskHandler)
	r.Put("/api/task", handlers.PutTaskHandler)
	r.Post("/api/task/done", handlers.DoneTaskHandler)
	r.Delete("/api/task", handlers.DeleteTaskHandler)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		log.Fatalf("Server crash: %v\n", err)
		return
	}
}
