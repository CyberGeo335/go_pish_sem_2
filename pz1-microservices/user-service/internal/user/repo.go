package user

import (
	"errors"
	"sort"
)

var ErrUserNotFound = errors.New("user not found")

// Repo is a simple in-memory user repository for the practice task.
type Repo struct {
	data map[int64]User
}

func NewRepo() *Repo {
	return &Repo{
		data: map[int64]User{
			1: {ID: 1, Name: "Иван Иванов", Email: "ivan@example.com"},
			2: {ID: 2, Name: "Мария Петрова", Email: "maria@example.com"},
			3: {ID: 3, Name: "Алексей Сидоров", Email: "alex@example.com"},
		},
	}
}

func (r *Repo) GetByID(id int64) (User, error) {
	u, ok := r.data[id]
	if !ok {
		return User{}, ErrUserNotFound
	}

	return u, nil
}

func (r *Repo) GetAll() []User {
	ids := make([]int64, 0, len(r.data))
	for id := range r.data {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	users := make([]User, 0, len(r.data))
	for _, id := range ids {
		users = append(users, r.data[id])
	}

	return users
}
