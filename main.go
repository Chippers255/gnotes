package main

import (
	"errors"
	"time"

	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	DB *gorm.DB
}

type Note struct {
	NoteID         string    `gorm:"primaryKey;column:note_id;type:uuid"`
	NoteTitle      string    `gorm:"column:note_title"`
	NoteContentURL string    `gorm:"column:note_content_url"`
	NoteCreatedAt  time.Time `gorm:"column:note_created_at"`
	NoteUpdatedAt  time.Time `gorm:"column:note_updated_at"`
	NoteLinks      []Link    `gorm:"foreignKey:LinkSourceNoteID"`
}

type Link struct {
	LinkID           string    `gorm:"primaryKey;column:link_id;type:uuid"`
	LinkSourceNoteID string    `gorm:"column:link_source_note_id;type:uuid"`
	LinkTargetNoteID string    `gorm:"column:link_target_note_id;type:uuid"`
	LinkSourceNote   Note      `gorm:"foreignKey:LinkSourceNoteID"`
	LinkTargetNote   Note      `gorm:"foreignKey:LinkTargetNoteID"`
	LinkCreatedAt    time.Time `gorm:"column:link_created_at"`
	LinkUpdatedAt    time.Time `gorm:"column:link_updated_at"`
}

func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("notes.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&Note{}, &Link{})
	if err != nil {
		panic("failed to migrate database")
	}

	return db
}

type NoteRequest struct {
	Title      string `json:"title"`
	ContentURL string `json:"content_url"`
}

func (app *App) CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	var req NoteRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	note := Note{
		NoteID:         uuid.New().String(),
		NoteTitle:      req.Title,
		NoteContentURL: req.ContentURL,
		NoteCreatedAt:  time.Now(),
		NoteUpdatedAt:  time.Now(),
	}

	result := app.DB.Create(&note)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func (app *App) GetNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID := vars["id"]

	var note Note
	result := app.DB.First(&note, "note_id = ?", noteID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(note)
}

func (app *App) ListNotesHandler(w http.ResponseWriter, r *http.Request) {
	var notes []Note
	result := app.DB.Find(&notes)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(notes)
}

func main() {
	db := initDB()
	app := &App{DB: db}

	r := mux.NewRouter()
	r.HandleFunc("/notes", app.CreateNoteHandler).Methods("POST")
	r.HandleFunc("/notes/{id}", app.GetNoteHandler).Methods("GET")
	r.HandleFunc("/notes", app.ListNotesHandler).Methods("GET")

	http.ListenAndServe(":8000", r)
}
