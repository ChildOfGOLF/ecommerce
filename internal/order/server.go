package order

import (
	"context"
	orderv1 "ecommerce/gen/order/v1"
	productv1 "ecommerce/gen/product/v1"
	userv1 "ecommerce/gen/user/v1"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	orderv1.UnimplementedOrderServiceServer
	store         *Store
	userClient    userv1.UserServiceClient
	productClient productv1.ProductServiceClient
}

func NewServer(store *Store, userClient userv1.UserServiceClient, productClient productv1.ProductServiceClient) *Server {
	return &Server{
		store:         store,
		userClient:    userClient,
		productClient: productClient,
	}
}

func (s *Server) CreateOrder(stream orderv1.OrderService_CreateOrderServer) error {
	ctx := stream.Context()

	var userID string
	var items []Item
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		userID = req.GetUserId()
		items = append(items, Item{
			ProductID: req.GetItem().GetProductId(),
			Quantity:  req.GetItem().GetQuantity(),
		})
	}

	if userID == "" || len(items) == 0 {
		return status.Error(codes.InvalidArgument, "user or item empty")
	}

	if _, err := s.userClient.GetUser(ctx, &userv1.GetUserRequest{UserId: userID}); err != nil {
		return status.Errorf(codes.FailedPrecondition, "user check failed: %v", err)
	}

	type reserved struct {
		productID string
		quantity  int32
	}

	var done []reserved

	for _, item := range items {
		_, err := s.productClient.UpdateStock(ctx, &productv1.UpdateStockRequest{
			ProductId: item.ProductID,
			Delta:     -item.Quantity,
		})
		if err != nil {
			// откат списанных
			for _, d := range done {
				// TODO: retry with backoff || outbox паттерн || очередь с гарантом
				// err ignore: если product service упадет навсегда рассинхрон
				_, _ = s.productClient.UpdateStock(ctx, &productv1.UpdateStockRequest{
					ProductId: d.productID,
					Delta:     d.quantity,
				})
			}
			return status.Errorf(codes.FailedPrecondition, "reserve stock for %s: %v", item.ProductID, err)
		}

		done = append(done, reserved{productID: item.ProductID, quantity: item.Quantity})
	}

	o, err := s.store.CreateOrder(userID, items)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return stream.SendAndClose(&orderv1.CreateOrderResponse{Order: toProto(o)})
}

func (s *Server) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	o, ok := s.store.GetOrder(req.GetOrderId())
	if !ok {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return &orderv1.GetOrderResponse{Order: toProto(o)}, nil
}

func (s *Server) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	orders := s.store.ListOrders()
	out := make([]*orderv1.Order, 0, len(orders))
	for _, o := range orders {
		out = append(out, toProto(o))
	}

	return &orderv1.ListOrdersResponse{Orders: out}, nil
}

func toProto(o Order) *orderv1.Order {
	// think about orderItemToProto
	items := make([]*orderv1.OrderItem, 0, len(o.Items))

	for _, item := range o.Items {
		items = append(items, &orderv1.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return &orderv1.Order{
		OrderId:   o.ID,
		UserId:    o.UserID,
		Items:     items,
		Status:    statusToProto(o.Status),
		CreatedAt: timestamppb.New(o.CreatedAt),
	}
}

func statusToProto(s OrderStatus) orderv1.OrderStatus {
	switch s {
	case OrderStatusPending:
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	case OrderStatusConfirmed:
		return orderv1.OrderStatus_ORDER_STATUS_CONFIRMED
	case OrderStatusCancelled:
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
