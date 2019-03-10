package todos

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

// Todo represents a single task
type Todo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	DueDate    string    `json:"due_date"`
	DueClock   string    `json:"due_clock"`
	DueTime    time.Time `json:"due_time"`
	NotifiedAt time.Time `json:"notified_at"`
}

// Manager manages todos
type Manager interface {
	Create(t Todo) error
	Read() ([]Todo, error)
	Update(t Todo) error
	Delete(t Todo) error
	Run(notifications chan<- Todo) error
}

// Manager manages todos
type manager struct {
	todos    []Todo
	lock     sync.Mutex
	savePath string
}

// NewManager returns an implementation of Manager
func NewManager(savePath string) (Manager, error) {
	todos, err := loadTodos(savePath)
	if err != nil {
		return nil, errors.Wrap(err, "loading todos failed")
	}
	return &manager{
		todos:    todos,
		savePath: savePath,
	}, nil
}

func saveTodos(todos []Todo, savePath string) error {
	f, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "opening file failed: %s", savePath)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(todos)
	return errors.Wrap(err, "encoding todos failed")
}

func loadTodos(savePath string) ([]Todo, error) {
	todos := []Todo{}
	if fi, err := os.Stat(savePath); os.IsNotExist(err) || fi.Size() == 0 {
		return todos, nil
	}
	f, err := os.OpenFile(savePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(f).Decode(&todos)
	return todos, errors.Wrap(err, "decoding todos failed")
}

func (m *manager) Create(t Todo) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	t, err := newTodo(t)
	if err != nil {
		return err
	}
	newTodos := append(m.todos, t)
	err = saveTodos(newTodos, m.savePath)
	if err != nil {
		return err
	}
	m.todos = newTodos
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
			mergo.Merge(&todo, t, mergo.WithOverride)
			var err error
			todo, err = setDateTime(todo, time.Now())
			if err != nil {
				return err
			}
		}
		newTodos = append(newTodos, todo)
	}
	err := saveTodos(newTodos, m.savePath)
	if err != nil {
		return err
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
	err := saveTodos(newTodos, m.savePath)
	if err != nil {
		return err
	}
	m.todos = newTodos
	return nil
}

func (m *manager) Run(notifications chan<- Todo) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.checkTodos(notifications)
		}
	}
}

func (m *manager) checkTodos(notifications chan<- Todo) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	currentTime := time.Now()
	todoUpdated := false
	newTodos := []Todo{}
	for _, todo := range m.todos {
		if todo.DueTime.Before(currentTime) &&
			currentTime.Sub(todo.NotifiedAt) > time.Duration(24)*time.Hour {
			notifications <- todo
			todo.NotifiedAt = currentTime
			todoUpdated = true
		}
		newTodos = append(newTodos, todo)
	}
	if todoUpdated {
		err := saveTodos(newTodos, m.savePath)
		if err != nil {
			return err
		}
		m.todos = newTodos
	}
	return nil
}
