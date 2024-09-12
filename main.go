package main

import (
	"finalProject/configs"
	"finalProject/internal/database"
	"finalProject/internal/handlers"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	var api handlers.API
	db, err := database.CreateDB()
	if err != nil {
		log.Fatal(err)
	}

	store := db.NewTaskStore(db.DB)
	api = api.NewAPI(store)

	log.Printf("Database start")
	r := chi.NewRouter()
	fs := http.FileServer(http.Dir(configs.WebDir))

	r.Handle("/*", fs)
	log.Printf("Loaded frontend from %s\n", configs.WebDir)
	log.Printf("Start server in port %d\n", configs.Port)

	r.Get("/api/nextdate", api.GetNextDate)
	r.Post("/api/task", api.PostTaskHandler)
	r.Get("/api/tasks", api.GetTasksHandler)
	r.Get("/api/task", api.GetTaskHandler)
	r.Put("/api/task", api.PutTaskHandler)
	r.Post("/api/task/done", api.DoneTaskHandler)
	r.Delete("/api/task", api.DeleteTaskHandler)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", configs.Port), r); err != nil {
		log.Fatalf("Server crash: %v\n", err)
		return
	}
}
