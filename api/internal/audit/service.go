package audit

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type Event struct {
	ActorUserID  *string
	Action       string
	TargetUserID *string
	Detail       *string
	IPAddress    *string
}

type Service struct {
	queries *queries.Queries
}

func NewService(q *queries.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) Record(ctx context.Context, event Event) error {
	if err := s.queries.CreateAuditLog(ctx, queries.CreateAuditLogParams{
		ActorUserID:  event.ActorUserID,
		Action:       event.Action,
		TargetUserID: event.TargetUserID,
		Detail:       event.Detail,
		IpAddress:    event.IPAddress,
	}); err != nil {
		return fmt.Errorf("record audit event %q: %w", event.Action, err)
	}
	return nil
}
