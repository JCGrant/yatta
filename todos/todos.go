package todos

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// Todo represents a single task
type Todo struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	DueDate  string    `json:"due_date"`
	DueClock string    `json:"due_clock"`
	DueTime  time.Time `json:"due_time"`
}

// Manager manages todos
type Manager interface {
	Create(t Todo) error
	Read() ([]Todo, error)
	Update(t Todo) error
	Delete(t Todo) error
}

// Manager manages todos
type manager struct {
	todos []Todo
	lock  sync.Mutex
}

// NewManager returns an implementation of Manager
func NewManager() Manager {
	return &manager{
		todos: []Todo{},
	}
}

func (m *manager) Create(t Todo) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	t, err := newTodo(t)
	if err != nil {
		return err
	}
	m.todos = append(m.todos, t)
	return nil
}

func newTodo(t Todo) (Todo, error) {
	id, err := newID()
	if err != nil {
		return Todo{}, err
	}
	t.ID = id
	t, err = setDateTime(t, time.Now())
	return t, err
}

func newID() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", errors.Wrap(err, "generating UUID failed")
	}
	return id.String(), nil
}

func setDateTime(t Todo, now time.Time) (Todo, error) {
	if t.DueDate == "" {
		t.DueDate = now.Format("2006-01-02")
	}
	if t.DueClock == "" {
		t.DueClock = "00:00:00"
	}
	dueTime, err := time.Parse("2006-01-02_15:04:05", fmt.Sprintf("%s_%s", t.DueDate, t.DueClock))
	t.DueTime = dueTime
	return t, err
}

func (m *manager) Read() ([]Todo, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.todos, nil
}

func (m *manager) Update(t Todo) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	newTodos := []Todo{}
	for _, todo := range m.todos {
		if todo.ID == t.ID {
			todo = t
		}
		newTodos = append(newTodos, todo)
	}
	m.todos = newTodos
	return nil
}

func (m *manager) Delete(t Todo) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	newTodos := []Todo{}
	for _, todo := range m.todos {
		if todo.ID == t.ID {
			continue
		}
		newTodos = append(newTodos, todo)
	}
	m.todos = newTodos
	return nil
}
