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
	var api handlers.API
	db, err := database.CreateDB()
	if err != nil {
		log.Fatal(err)
	}

	store := db.NewTaskStore(db.DB)
	api = api.NewAPI(store)

	log.Printf("Database start")
	r := chi.NewRouter()
	fs := http.FileServer(http.Dir(webDir))

	r.Handle("/*", fs)
	log.Printf("Loaded frontend from %s\n", webDir)
	log.Printf("Start server in port %d\n", port)

	r.Get("/api/nextdate", api.GetNextDate)
	r.Post("/api/task", api.PostTaskHandler)
	r.Get("/api/tasks", api.GetTasksHandler)
	r.Get("/api/task", api.GetTaskHandler)
	r.Put("/api/task", api.PutTaskHandler)
	r.Post("/api/task/done", api.DoneTaskHandler)
	r.Delete("/api/task", api.DeleteTaskHandler)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		log.Fatalf("Server crash: %v\n", err)
		return
	}
}
