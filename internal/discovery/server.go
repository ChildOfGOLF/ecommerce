package discovery

import (
	"context"
	discoveryv1 "ecommerce/gen/discovery/v1"
	"io"
	"time"
)

const heartbeatTTL = 15 * time.Second

type Server struct {
	discoveryv1.UnimplementedDiscoveryServiceServer
	store *Store
}

func NewServer(store *Store) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) Resolve(ctx context.Context, req *discoveryv1.ResolveRequest) (*discoveryv1.ResolveResponse, error) {
	instances := s.store.Resolve(req.GetServiceName(), heartbeatTTL)

	result := make([]*discoveryv1.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		result = append(result, toProto(inst))
	}

	return &discoveryv1.ResolveResponse{
		Instances: result,
	}, nil
}

func (s *Server) Heartbeat(stream discoveryv1.DiscoveryService_HeartbeatServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		inst := req.GetInstance()

		s.store.Upsert(Instance{
			ServiceName: inst.GetServiceName(),
			InstanceID:  inst.GetInstanceId(),
			Address:     inst.GetAddress(),
		})

		if err := stream.Send(&discoveryv1.HeartbeatResponse{
			Acknowledged:    true,
			IntervalSeconds: 10,
		}); err != nil {
			return err
		}
	}
}

func toProto(inst Instance) *discoveryv1.ServiceInstance {
	return &discoveryv1.ServiceInstance{
		ServiceName: inst.ServiceName,
		InstanceId:  inst.InstanceID,
		Address:     inst.Address,
	}
}
