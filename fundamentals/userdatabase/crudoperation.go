package userdatabase

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Student struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type StudentHandler struct {
	db *sql.DB
}

func NewStudentHandler(db *sql.DB) *StudentHandler {
	return &StudentHandler{
		db: db,
	}
}
func (s *StudentHandler) GetStudents(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id , name , email FROM students")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.Name, &s.Email); err != nil {
			http.Error(w, "rows scan failed", http.StatusInternalServerError)
			return
		}
		students = append(students, s)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}
func (s *StudentHandler) InsertStudents(w http.ResponseWriter, r *http.Request) {
	var Newstudent Student
	if err := json.NewDecoder(r.Body).Decode(&Newstudent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := s.db.Exec("INSERT INTO students(name , email)VALUES (? , ?)", Newstudent.Name, Newstudent.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	Newstudent.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Newstudent)
}
func (s *StudentHandler) UpdateStudents(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	var updated Student
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err := s.db.Exec("UPDATE students  SET name=?, email=? WHERE id=? ", updated.Name, updated.Email, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
func (s *StudentHandler) DeleteStudents(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	_, err := s.db.Exec("DELETE FROM students WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (s *StudentHandler) getstudentsbyId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var student Student
	err = s.db.QueryRow("Select id , name , email FROM students WHERE id=?", id).
		Scan(&student.ID, &student.Name, &student.Email)

	if err == sql.ErrNoRows {
		http.Error(w, "student not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)

}
func (s *StudentHandler) PatchStudents(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	// Decode incoming JSON into a map so we can handle partial updates
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build dynamic SQL based on provided fields
	query := "UPDATE students SET "
	args := []interface{}{}
	first := true

	if name, ok := updates["name"]; ok {
		if !first {
			query += ", "
		}
		query += "name=?"
		args = append(args, name)
		first = false
	}

	if email, ok := updates["email"]; ok {
		if !first {
			query += ", "
		}
		query += "email=?"
		args = append(args, email)
		first = false
	}

	query += " WHERE id=?"
	args = append(args, id)

	_, err = s.db.Exec(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated student object
	var student Student
	err = s.db.QueryRow("SELECT id, name, email FROM students WHERE id=?", id).
		Scan(&student.ID, &student.Name, &student.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "student not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

func CRUDOperation() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("err loading .env file: %v", err)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	fmt.Println("dsn:", dsn)
	var err error
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping error:%v", err)
	}
	fmt.Println("connection succesfull!")

	stdhandler := NewStudentHandler(db)
	r := mux.NewRouter()
	r.HandleFunc("/students", stdhandler.GetStudents).Methods("GET")
	r.HandleFunc("/students", stdhandler.InsertStudents).Methods("POST")
	r.HandleFunc("/students/{id}", stdhandler.UpdateStudents).Methods("PUT")
	r.HandleFunc("/students/{id}", stdhandler.DeleteStudents).Methods("DELETE")
	r.HandleFunc("/students/{id}", stdhandler.getstudentsbyId).Methods("GET")
	r.HandleFunc("/students/{id}", stdhandler.PatchStudents).Methods("PATCH")

	fmt.Println("server running on port:8080")
	http.ListenAndServe(":8080", r)
}
