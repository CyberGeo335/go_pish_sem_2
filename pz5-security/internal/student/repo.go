package student

import (
	"database/sql"
	"errors"
)

var ErrStudentNotFound = errors.New("student not found")

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// UnsafeGetByID is an intentionally dangerous example.
// It is kept only to demonstrate why SQL string concatenation must not be used.
func (r *Repo) UnsafeGetByID(rawID string) (*Student, error) {
	query := "SELECT id, full_name, study_group, email FROM students WHERE id = " + rawID

	row := r.db.QueryRow(query)

	var st Student
	err := row.Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	return &st, nil
}

func (r *Repo) GetByID(id int64) (*Student, error) {
	row := r.db.QueryRow(
		"SELECT id, full_name, study_group, email FROM students WHERE id = $1",
		id,
	)

	return scanStudent(row)
}

func (r *Repo) GetByEmail(email string) (*Student, error) {
	row := r.db.QueryRow(
		"SELECT id, full_name, study_group, email FROM students WHERE email = $1",
		email,
	)

	return scanStudent(row)
}

func (r *Repo) PrepareGetByID() (*sql.Stmt, error) {
	return r.db.Prepare("SELECT id, full_name, study_group, email FROM students WHERE id = $1")
}

func (r *Repo) PrepareGetByEmail() (*sql.Stmt, error) {
	return r.db.Prepare("SELECT id, full_name, study_group, email FROM students WHERE email = $1")
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanStudent(row rowScanner) (*Student, error) {
	var st Student
	err := row.Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	return &st, nil
}
