package user

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           string
	Email        string
	FullName     string
	PasswordHash string
	CreatedAt    time.Time
}

type Store struct {
	// TODO: sync.RWMutex
	mu    sync.Mutex
	users map[string]User
	// TODO: usersByEmail map[string]string
}

func NewStore() *Store {
	return &Store{
		users: make(map[string]User),
	}
}

func (s *Store) Create(email, fullName, passwordHash string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Email == email {
			return User{}, errors.New("user already exists")
		}
	}

	user := User{
		ID:           uuid.NewString(),
		Email:        email,
		FullName:     fullName,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}

	s.users[user.ID] = user

	return user, nil
}

func (s *Store) Get(id string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]

	return user, ok
}

func (s *Store) List() []User {
	s.mu.Lock()
	defer s.mu.Unlock()

	users := make([]User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}

	return users
}
