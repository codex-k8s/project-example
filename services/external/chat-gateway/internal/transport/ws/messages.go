package ws

import (
	"encoding/json"
	"fmt"

	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/service"
)

type Envelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func EncodeEvent(e service.Event) ([]byte, error) {
	switch e.Type {
	case service.EventMessageCreated:
		if e.Message == nil {
			return nil, fmt.Errorf("ws event invariant: message is nil")
		}
		return json.Marshal(Envelope{
			Type:    string(e.Type),
			Payload: map[string]any{"message": e.Message},
		})
	case service.EventMessageDeleted:
		return json.Marshal(Envelope{
			Type: string(e.Type),
			Payload: map[string]any{
				"message_id": e.MessageID,
				"deleted_at": e.DeletedAt,
			},
		})
	default:
		return nil, fmt.Errorf("unknown event type: %q", e.Type)
	}
}
