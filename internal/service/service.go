package service

import (
	"context"
	"sync"

	"vk_task/proto"                  
	"vk_task/pkg/subpub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PubSubService struct {
	proto.UnimplementedPubSubServer
	bus subpub.SubPub
	mu  sync.Mutex
}

func NewPubSubService(bus subpub.SubPub) *PubSubService {
	return &PubSubService{
		bus: bus,
	}
}

func (s *PubSubService) Subscribe(req *proto.SubscribeRequest, stream proto.PubSub_SubscribeServer) error {
	ctx := stream.Context()

	sub, err := s.bus.Subscribe(req.Key, func(msg interface{}) {
		data, ok := msg.(string)
		if !ok {
			return
		}

		if err := stream.Send(&proto.Event{Data: data}); err != nil {
			return
		}
	})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	<-ctx.Done()
	return nil
}

func (s *PubSubService) Publish(ctx context.Context, req *proto.PublishRequest) (*emptypb.Empty, error) {
	if err := s.bus.Publish(req.Key, req.Data); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish: %v", err)
	}
	return &emptypb.Empty{}, nil
}