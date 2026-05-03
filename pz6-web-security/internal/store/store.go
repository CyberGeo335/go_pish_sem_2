package store

import (
	"sync"
	"time"
)

type UserProfile struct {
	SessionID string
	Name      string
	CSRFToken string
}

type Comment struct {
	Author    string
	Text      string
	CreatedAt time.Time
}

type Store struct {
	mu       sync.RWMutex
	users    map[string]*UserProfile
	comments []Comment
}

func New() *Store {
	return &Store{
		users:    make(map[string]*UserProfile),
		comments: make([]Comment, 0),
	}
}

func (s *Store) Save(profile *UserProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[profile.SessionID] = profile
}

func (s *Store) Get(sessionID string) (*UserProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, ok := s.users[sessionID]
	return profile, ok
}

func (s *Store) UpdateName(sessionID, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile, ok := s.users[sessionID]
	if !ok {
		return false
	}

	profile.Name = name
	return true
}

func (s *Store) UpdateCSRFToken(sessionID, csrfToken string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile, ok := s.users[sessionID]
	if !ok {
		return false
	}

	profile.CSRFToken = csrfToken
	return true
}

func (s *Store) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, sessionID)
}

func (s *Store) AddComment(comment Comment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.comments = append(s.comments, comment)
}

func (s *Store) Comments() []Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Comment, len(s.comments))
	copy(result, s.comments)
	return result
}
