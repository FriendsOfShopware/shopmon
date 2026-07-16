package audit

import (
	"context"
	"fmt"
)

type Event struct {
	ActorUserID  *string
	Action       string
	TargetUserID *string
	Detail       *string
	IPAddress    *string
}

type Repository interface {
	Create(ctx context.Context, event Event) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Record(ctx context.Context, event Event) error {
	if err := s.repository.Create(ctx, event); err != nil {
		return fmt.Errorf("record audit event %q: %w", event.Action, err)
	}
	return nil
}
