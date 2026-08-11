package order

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID        string
	UserID    string
	Items     []Item
	Status    OrderStatus
	CreatedAt time.Time
}

type Item struct {
	ProductID string
	Quantity  int32
}

type Store struct {
	// TODO: sync.RWMutex
	mu     sync.Mutex
	orders map[string]Order
}

func NewStore() *Store {
	return &Store{
		orders: make(map[string]Order),
	}
}

func (s *Store) CreateOrder(userID string, items []Item) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order := Order{
		ID:        uuid.NewString(),
		UserID:    userID,
		Items:     items,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	s.orders[order.ID] = order

	return order, nil
}

func (s *Store) GetOrder(id string) (Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[id]

	return order, ok
}

func (s *Store) ListOrders() []Order {
	s.mu.Lock()
	defer s.mu.Unlock()

	orders := make([]Order, 0, len(s.orders))
	for _, o := range s.orders {
		orders = append(orders, o)
	}

	return orders
}
