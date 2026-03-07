package persona

import (
	"context"
	"fmt"
	"log/slog"
)

// Service provides UI-facing read APIs for persona data.
type Service struct {
	store  PersonaStore
	logger *slog.Logger
}

// NewService creates a persona service backed by PersonaStore.
func NewService(store PersonaStore, logger *slog.Logger) *Service {
	return &Service{
		store:  store,
		logger: logger.With("slice", "PersonaService"),
	}
}

// ListNPCs returns persona NPC rows for UI table rendering.
func (s *Service) ListNPCs() ([]PersonaNPCView, error) {
	rows, err := s.store.ListNPCs(context.Background())
	if err != nil {
		s.logger.Error("failed to list persona npcs", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to list persona npcs: %w", err)
	}
	return rows, nil
}

// ListDialoguesBySpeaker returns dialogues for one speaker.
func (s *Service) ListDialoguesBySpeaker(speakerID string) ([]PersonaDialogueView, error) {
	rows, err := s.store.ListDialoguesBySpeaker(context.Background(), speakerID)
	if err != nil {
		s.logger.Error("failed to list persona dialogues", slog.String("speaker_id", speakerID), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to list persona dialogues: %w", err)
	}
	return rows, nil
}
