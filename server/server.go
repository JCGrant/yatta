package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JCGrant/yatta/notifiers"
	"github.com/JCGrant/yatta/todos"
)

// Server serves
type Server struct {
	port        int
	todoManager todos.Manager
	notifiers   []notifiers.Notifier
}

// New creates a Server
func New(port int, todoManager todos.Manager, notifiers ...notifiers.Notifier) *Server {
	return &Server{
		port:        port,
		todoManager: todoManager,
		notifiers:   notifiers,
	}
}

// Start starts the Server
func (s *Server) Start() error {
	http.HandleFunc("/todo", s.handleTodos)
	log.Println(fmt.Sprintf("running on http://127.0.0.1:%d", s.port))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	methods := map[string]http.HandlerFunc{
		http.MethodPost:   s.createTodo,
		http.MethodGet:    s.readTodos,
		http.MethodPut:    s.updateTodo,
		http.MethodDelete: s.deleteTodo,
	}
	handler, ok := methods[r.Method]
	if !ok {
		sendError(w, "unsupported method")
		return
	}
	handler(w, r)
}

func (s *Server) createTodo(w http.ResponseWriter, r *http.Request) {
	todo, err := getTodoFromRequest(r)
	if err != nil {
		sendError(w, err)
		return
	}
	err = s.todoManager.Create(todo)
	if err != nil {
		sendError(w, err)
	}
}

func (s *Server) readTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := s.todoManager.Read()
	if err != nil {
		sendError(w, err)
		return
	}
	sendResponse(w, todos)
	return
}

func (s *Server) updateTodo(w http.ResponseWriter, r *http.Request) {
	todo, err := getTodoFromRequest(r)
	if err != nil {
		sendError(w, err)
		return
	}
	err = s.todoManager.Update(todo)
	if err != nil {
		sendError(w, err)
		return
	}
}

func (s *Server) deleteTodo(w http.ResponseWriter, r *http.Request) {
	todo, err := getTodoFromRequest(r)
	if err != nil {
		sendError(w, err)
		return
	}
	err = s.todoManager.Delete(todo)
	if err != nil {
		sendError(w, err)
		return
	}
}

func getTodoFromRequest(r *http.Request) (todos.Todo, error) {
	var todo todos.Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	return todo, err
}

type serverResponse struct {
	Value interface{} `json:"value"`
	Time  time.Time   `json:"time"`
}

func sendResponse(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(serverResponse{
		Value: v,
		Time:  time.Now(),
	})
}

func sendError(w http.ResponseWriter, v interface{}) {
	w.WriteHeader(500)
	fmt.Fprint(w, v)
}
