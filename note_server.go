package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
)

type Note struct {
    ID      int    `json:"id"`
    Content string `json:"content"`
}

type NoteStore struct {
    mu    sync.Mutex
    notes map[int]Note
    nextID int
}

func NewNoteStore() *NoteStore {
    return &NoteStore{
        notes: make(map[int]Note),
    }
}

func (store *NoteStore) Add(content string) int {
    store.mu.Lock()
    defer store.mu.Unlock()
    note := Note{ID: store.nextID, Content: content}
    store.notes[store.nextID] = note
    store.nextID++
    return note.ID
}

func (store *NoteStore) GetAll() []Note {
    store.mu.Lock()
    defer store.mu.Unlock()
    notes := make([]Note, 0, len(store.notes))
    for _, note := range store.notes {
        notes = append(notes, note)
    }
    return notes
}

func (store *NoteStore) Update(id int, content string) bool {
    store.mu.Lock()
    defer store.mu.Unlock()
    if note, exists := store.notes[id]; exists {
        note.Content = content
        store.notes[id] = note
        return true
    }
    return false
}

func (store *NoteStore) Delete(id int) bool {
    store.mu.Lock()
    defer store.mu.Unlock()
    if _, exists := store.notes[id]; exists {
        delete(store.notes, id)
        return true
    }
    return false
}

var noteStore = NewNoteStore()

func main() {
    http.HandleFunc("/notes", notesHandler)
    http.HandleFunc("/notes/", noteHandler)
    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil)
}

func notesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        notes := noteStore.GetAll()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(notes)
    case http.MethodPost:
        var note Note
        if err := json.NewDecoder(r.Body).Decode(&note); err == nil {
            id := noteStore.Add(note.Content)
            w.WriteHeader(http.StatusCreated)
            json.NewEncoder(w).Encode(map[string]int{"id": id})
        } else {
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func noteHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/notes/"):]

    switch r.Method {
    case http.MethodPut:
        var note Note
        if err := json.NewDecoder(r.Body).Decode(&note); err == nil {
            if noteStore.Update(id, note.Content) {
                w.WriteHeader(http.StatusNoContent)
            } else {
                http.Error(w, "Note not found", http.StatusNotFound)
            }
        } else {
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
    case http.MethodDelete:
        if noteStore.Delete(id) {
            w.WriteHeader(http.StatusNoContent)
        } else {
            http.Error(w, "Note not found", http.StatusNotFound)
        }
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
