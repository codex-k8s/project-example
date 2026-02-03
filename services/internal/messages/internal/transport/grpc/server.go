package grpc

import (
	"context"
	"fmt"

	msgsvc "github.com/codex-k8s/project-example/services/internal/messages/internal/domain/service"
	"github.com/codex-k8s/project-example/services/internal/messages/internal/domain/types/entity"
	grpcgen "github.com/codex-k8s/project-example/services/internal/messages/internal/transport/grpc/generated/messages/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	grpcgen.UnimplementedMessagesServiceServer
	svc *msgsvc.Service
}

func Register(s *grpc.Server, svc *msgsvc.Service) {
	grpcgen.RegisterMessagesServiceServer(s, &Server{svc: svc})
}

func (s *Server) CreateMessage(ctx context.Context, req *grpcgen.CreateMessageRequest) (*grpcgen.CreateMessageResponse, error) {
	msg, err := s.svc.CreateMessage(ctx, req.GetUserId(), req.GetText())
	if err != nil {
		return nil, err
	}
	return &grpcgen.CreateMessageResponse{Message: toProtoMessage(msg)}, nil
}

func (s *Server) DeleteMessage(ctx context.Context, req *grpcgen.DeleteMessageRequest) (*grpcgen.DeleteMessageResponse, error) {
	msg, err := s.svc.DeleteMessage(ctx, req.GetUserId(), req.GetMessageId())
	if err != nil {
		return nil, err
	}
	if msg.DeletedAt == nil {
		return nil, fmt.Errorf("invariant: deleted_at is nil after delete")
	}
	return &grpcgen.DeleteMessageResponse{
		MessageId: msg.ID,
		DeletedAt: timestamppb.New(*msg.DeletedAt),
	}, nil
}

func (s *Server) ListMessages(ctx context.Context, req *grpcgen.ListMessagesRequest) (*grpcgen.ListMessagesResponse, error) {
	msgs, err := s.svc.ListRecent(ctx, int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	out := make([]*grpcgen.Message, 0, len(msgs))
	for i := range msgs {
		out = append(out, toProtoMessage(msgs[i]))
	}
	return &grpcgen.ListMessagesResponse{Messages: out}, nil
}

func (s *Server) PurgeOldMessages(ctx context.Context, req *grpcgen.PurgeOldMessagesRequest) (*grpcgen.PurgeOldMessagesResponse, error) {
	ts := req.GetOlderThan()
	if ts == nil {
		return nil, fmt.Errorf("older_than is required")
	}
	olderThan := ts.AsTime()
	deleted, err := s.svc.PurgeOld(ctx, olderThan)
	if err != nil {
		return nil, err
	}
	out := make([]*grpcgen.DeleteMessageResponse, 0, len(deleted))
	for i := range deleted {
		m := deleted[i]
		if m.DeletedAt == nil {
			continue
		}
		out = append(out, &grpcgen.DeleteMessageResponse{
			MessageId: m.ID,
			DeletedAt: timestamppb.New(*m.DeletedAt),
		})
	}
	return &grpcgen.PurgeOldMessagesResponse{Deleted: out}, nil
}

func (s *Server) SubscribeEvents(req *grpcgen.SubscribeEventsRequest, stream grpcgen.MessagesService_SubscribeEventsServer) error {
	_ = req
	ch := s.svc.Subscribe(stream.Context())
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case evt, ok := <-ch:
			if !ok {
				return nil
			}
			msg, err := toProtoEvent(evt)
			if err != nil {
				return err
			}
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}
}

func toProtoMessage(m entity.Message) *grpcgen.Message {
	created := timestamppb.New(m.CreatedAt)
	if err := created.CheckValid(); err != nil {
		panic(fmt.Errorf("invalid created_at: %w", err))
	}

	var deleted *timestamppb.Timestamp
	if m.DeletedAt != nil {
		ts := timestamppb.New(*m.DeletedAt)
		if err := ts.CheckValid(); err != nil {
			panic(fmt.Errorf("invalid deleted_at: %w", err))
		}
		deleted = ts
	}

	return &grpcgen.Message{
		Id:        m.ID,
		UserId:    m.UserID,
		Text:      m.Text,
		CreatedAt: created,
		DeletedAt: deleted,
	}
}

func toProtoEvent(e msgsvc.Event) (*grpcgen.Event, error) {
	switch e.Type {
	case msgsvc.EventMessageCreated:
		if e.Message == nil {
			return nil, fmt.Errorf("event invariant: message is nil")
		}
		return &grpcgen.Event{
			Payload: &grpcgen.Event_MessageCreated{
				MessageCreated: &grpcgen.MessageCreated{Message: toProtoMessage(*e.Message)},
			},
		}, nil
	case msgsvc.EventMessageDeleted:
		if e.MessageID <= 0 || e.DeletedAt == nil {
			return nil, fmt.Errorf("event invariant: message_id/deleted_at invalid")
		}
		return &grpcgen.Event{
			Payload: &grpcgen.Event_MessageDeleted{
				MessageDeleted: &grpcgen.MessageDeleted{
					MessageId: e.MessageID,
					DeletedAt: timestamppb.New(*e.DeletedAt),
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown event type: %q", string(e.Type))
	}
}
