package registry

import (
	"context"
	discoveryv1 "ecommerce/gen/discovery/v1"
	"log"
	"time"

	"github.com/google/uuid"
)

func RegisterAndHeartbeat(ctx context.Context, client discoveryv1.DiscoveryServiceClient, serviceName, address string) {
	instanceID := uuid.NewString()

	stream, err := client.Heartbeat(ctx)
	if err != nil {
		log.Print(err)
		return
	}

	send := func() {
		// TODO: reconnect with backlog
		// если упадет send() продолжит вызываться
		err := stream.Send(&discoveryv1.HeartbeatRequest{
			Instance: &discoveryv1.ServiceInstance{
				ServiceName: serviceName,
				InstanceId:  instanceID,
				Address:     address,
			},
		})
		if err != nil {
			log.Print(err)
		}
	}

	send()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		send()
	}
}
