package product

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

type Product struct {
	ID    string
	Name  string
	Price int64
	Stock int32
}

type Store struct {
	// TODO: sync.RWMutex
	mu          sync.Mutex
	products    map[string]Product
	subscribers map[chan StockUpdate]string
}

type StockUpdate struct {
	ProductID string
	Stock     int32
}

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

func NewStore() *Store {
	return &Store{
		products:    make(map[string]Product),
		subscribers: make(map[chan StockUpdate]string),
	}
}

func (s *Store) CreateProduct(name string, price int64, stock int32) (Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	product := Product{
		ID:    uuid.NewString(),
		Name:  name,
		Price: price,
		Stock: stock,
	}

	s.products[product.ID] = product

	return product, nil
}

func (s *Store) GetProduct(id string) (Product, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	product, ok := s.products[id]

	return product, ok
}

func (s *Store) ListProducts() []Product {
	s.mu.Lock()
	defer s.mu.Unlock()

	products := make([]Product, 0, len(s.products))
	for _, p := range s.products {
		products = append(products, p)
	}

	return products
}

func (s *Store) Subscribe(productID string) chan StockUpdate {
	ch := make(chan StockUpdate, 1)
	s.mu.Lock()
	s.subscribers[ch] = productID
	s.mu.Unlock()
	return ch
}

func (s *Store) Unsubscribe(ch chan StockUpdate) {
	s.mu.Lock()
	delete(s.subscribers, ch)
	close(ch)
	s.mu.Unlock()
}

func (s *Store) UpdateStock(productID string, delta int32) (Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	product, ok := s.products[productID]
	if !ok {
		return Product{}, ErrProductNotFound
	}

	newStock := product.Stock + delta
	if newStock < 0 {
		return Product{}, ErrInsufficientStock
	}

	product.Stock = newStock
	s.products[productID] = product

	event := StockUpdate{ProductID: productID, Stock: newStock}
	for ch, filter := range s.subscribers {
		if filter != "" && filter != productID {
			continue
		}
		select {
		case ch <- event:
		default:
		}
	}
	return product, nil
}
