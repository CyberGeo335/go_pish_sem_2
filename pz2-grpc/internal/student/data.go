package student

import (
	"errors"
	"sort"
	"sync"

	"github.com/CyberGeo335/pz2-grpc/gen/studentpb"
)

var ErrStudentNotFound = errors.New("student not found")

type Repository struct {
	mu     sync.RWMutex
	data   map[int64]*studentpb.Student
	nextID int64
}

func NewRepository() *Repository {
	return &Repository{
		data: map[int64]*studentpb.Student{
			1: {
				Id:             1,
				FullName:       "Иванов Иван Иванович",
				Group:          "ИВБО-01-25",
				Email:          "ivanov@example.com",
				Specialization: "Backend-разработка",
			},
			2: {
				Id:             2,
				FullName:       "Петрова Мария Сергеевна",
				Group:          "ИВБО-02-25",
				Email:          "petrova@example.com",
				Specialization: "DevOps",
			},
			3: {
				Id:             3,
				FullName:       "Сидоров Алексей Андреевич",
				Group:          "ИВБО-03-25",
				Email:          "sidorov@example.com",
				Specialization: "Информационная безопасность",
			},
		},
		nextID: 4,
	}
}

func (r *Repository) GetByID(id int64) (*studentpb.Student, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	st, ok := r.data[id]
	if !ok {
		return nil, ErrStudentNotFound
	}

	return cloneStudent(st), nil
}

func (r *Repository) List() []*studentpb.Student {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]int64, 0, len(r.data))
	for id := range r.data {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	students := make([]*studentpb.Student, 0, len(ids))
	for _, id := range ids {
		students = append(students, cloneStudent(r.data[id]))
	}

	return students
}

func (r *Repository) Create(fullName, group, email, specialization string) *studentpb.Student {
	r.mu.Lock()
	defer r.mu.Unlock()

	st := &studentpb.Student{
		Id:             r.nextID,
		FullName:       fullName,
		Group:          group,
		Email:          email,
		Specialization: specialization,
	}

	r.data[st.Id] = st
	r.nextID++

	return cloneStudent(st)
}

func cloneStudent(st *studentpb.Student) *studentpb.Student {
	if st == nil {
		return nil
	}
	copy := *st
	return &copy
}
