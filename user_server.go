package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

type UserStore struct {
    mu     sync.Mutex
    users  map[int]User
    nextID int
}

func NewUserStore() *UserStore {
    return &UserStore{
        users: make(map[int]User),
    }
}

func (store *UserStore) Add(name string) int {
    store.mu.Lock()
    defer store.mu.Unlock()
    user := User{ID: store.nextID, Name: name, CreatedAt: time.Now()}
    store.users[store.nextID] = user
    store.nextID++
    return user.ID
}

func (store *UserStore) GetAll() []User {
    store.mu.Lock()
    defer store.mu.Unlock()
    users := make([]User, 0, len(store.users))
    for _, user := range store.users {
        users = append(users, user)
    }
    return users
}

func (store *UserStore) Delete(id int) bool {
    store.mu.Lock()
    defer store.mu.Unlock()
    if _, exists := store.users[id]; exists {
        delete(store.users, id)
        return true
    }
    return false
}

var userStore = NewUserStore()

func main() {
    http.HandleFunc("/users", usersHandler)
    http.HandleFunc("/users/", userHandler)
    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        users := userStore.GetAll()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(users)
    case http.MethodPost:
        var user User
        if err := json.NewDecoder(r.Body).Decode(&user); err == nil {
            id := userStore.Add(user.Name)
            w.WriteHeader(http.StatusCreated)
            json.NewEncoder(w).Encode(map[string]int{"id": id})
        } else {
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/users/"):]

    switch r.Method {
    case http.MethodDelete:
        if userStore.Delete(id) {
            w.WriteHeader(http.StatusNoContent)
        } else {
            http.Error(w, "User not found", http.StatusNotFound)
        }
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
