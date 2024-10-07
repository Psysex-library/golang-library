package main

import (
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "sync"
)

type Task struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Status string `json:"status"`
}

type TaskStore struct {
    mu    sync.Mutex
    tasks map[int]Task
    nextID int
}

func NewTaskStore() *TaskStore {
    return &TaskStore{
        tasks: make(map[int]Task),
    }
}

func (store *TaskStore) Add(name string) int {
    store.mu.Lock()
    defer store.mu.Unlock()
    task := Task{ID: store.nextID, Name: name, Status: "pending"}
    store.tasks[store.nextID] = task
    store.nextID++
    return task.ID
}

func (store *TaskStore) GetAll() []Task {
    store.mu.Lock()
    defer store.mu.Unlock()
    tasks := make([]Task, 0, len(store.tasks))
    for _, task := range store.tasks {
        tasks = append(tasks, task)
    }
    return tasks
}

func (store *TaskStore) Execute(id int) bool {
    store.mu.Lock()
    defer store.mu.Unlock()
    if task, exists := store.tasks[id]; exists {
        cmd := exec.Command("echo", task.Name)
        if err := cmd.Run(); err == nil {
            task.Status = "completed"
            store.tasks[id] = task
            return true
        }
    }
    return false
}

var taskStore = NewTaskStore()

func main() {
    http.HandleFunc("/tasks", tasksHandler)
    http.HandleFunc("/tasks/", taskHandler)
    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        tasks := taskStore.GetAll()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(tasks)
    case http.MethodPost:
        var task Task
        if err := json.NewDecoder(r.Body).Decode(&task); err == nil {
            id := taskStore.Add(task.Name)
            w.WriteHeader(http.StatusCreated)
            json.NewEncoder(w).Encode(map[string]int{"id": id})
        } else {
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/tasks/"):]

    switch r.Method {
    case http.MethodPost:
        if taskStore.Execute(id) {
            w.WriteHeader(http.StatusNoContent)
        } else {
            http.Error(w, "Task not found", http.StatusNotFound)
        }
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
