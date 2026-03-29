package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/sri-shubham/neon"
)

// Task represents a simple data model
type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// TaskService manages a list of tasks in memory
type TaskService struct {
	neon.Module `base:"/tasks"`
	mu          sync.RWMutex
	tasks       map[string]Task

	// Endpoints
	getTasks   neon.Get    `url:"/"`
	getTask    neon.Get    `url:"/{id}"`
	createTask neon.Post   `url:"/"`
	updateTask neon.Put    `url:"/{id}"`
	deleteTask neon.Delete `url:"/{id}"`
}

// Init initializes the task storage
func (s *TaskService) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = make(map[string]Task)
	s.tasks["1"] = Task{ID: "1", Title: "Learn Neon Framework", Completed: false}
	s.tasks["2"] = Task{ID: "2", Title: "Build something awesome", Completed: false}
	return nil
}

// GetTasks returns all tasks
func (s *TaskService) GetTasks(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTask returns a single task by ID
func (s *TaskService) GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.RLock()
	task, ok := s.tasks[id]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// CreateTask adds a new task
func (s *TaskService) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		task.ID = fmt.Sprintf("%d", len(s.tasks)+1)
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// UpdateTask updates an existing task
func (s *TaskService) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var updatedTask Task
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	if _, ok := s.tasks[id]; !ok {
		s.mu.Unlock()
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	updatedTask.ID = id
	s.tasks[id] = updatedTask
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}

// DeleteTask removes a task
func (s *TaskService) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.Lock()
	delete(s.tasks, id)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	app := neon.New()
	app.AddService(&TaskService{})

	// Setup context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("Neon CRUD API Example running at http://localhost:8080")
	fmt.Println("Try these commands:")
	fmt.Println("  curl http://localhost:8080/tasks")
	fmt.Println("  curl http://localhost:8080/tasks/1")
	fmt.Println("  curl -X POST -d '{\"title\":\"New Task\"}' http://localhost:8080/tasks")

	if err := app.Run(ctx); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
