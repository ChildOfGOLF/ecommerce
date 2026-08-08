package product

import (
	"context"
	productv1 "ecommerce/gen/product/v1"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	productv1.UnimplementedProductServiceServer
	store *Store
}

func NewServer(store *Store) *Server {
	return &Server{store: store}
}

func (s *Server) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	if req.GetPrice() < 0 {
		return nil, status.Error(codes.InvalidArgument, "price negative")
	}

	if req.GetStock() < 0 {
		return nil, status.Error(codes.InvalidArgument, "stock negative")
	}

	p, err := s.store.CreateProduct(req.GetName(), req.GetPrice(), req.GetStock())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productv1.CreateProductResponse{Product: toProto(p)}, nil
}

func (s *Server) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	p, ok := s.store.GetProduct(req.GetProductId())
	if !ok {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	return &productv1.GetProductResponse{Product: toProto(p)}, nil
}

func (s *Server) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	products := s.store.ListProducts()
	out := make([]*productv1.Product, 0, len(products))
	for _, p := range products {
		out = append(out, toProto(p))
	}
	return &productv1.ListProductsResponse{Products: out}, nil
}

func (s *Server) UpdateStock(ctx context.Context, req *productv1.UpdateStockRequest) (*productv1.UpdateStockResponse, error) {
	p, err := s.store.UpdateStock(
		req.GetProductId(),
		req.GetDelta(),
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			return nil, status.Error(codes.NotFound, err.Error())

		case errors.Is(err, ErrInsufficientStock):
			return nil, status.Error(codes.FailedPrecondition, err.Error())

		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &productv1.UpdateStockResponse{Product: toProto(p)}, nil
}

func (s *Server) WatchStock(req *productv1.WatchStockRequest, stream productv1.ProductService_WatchStockServer) error {
	ch := s.store.Subscribe(req.GetProductId())
	defer s.store.Unsubscribe(ch)

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case update, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(&productv1.StockUpdate{
				ProductId: update.ProductID,
				Stock:     update.Stock,
			}); err != nil {
				return err
			}
		}
	}
}

func toProto(p Product) *productv1.Product {
	return &productv1.Product{
		ProductId: p.ID,
		Name:      p.Name,
		Price:     p.Price,
		Stock:     p.Stock,
	}
}
