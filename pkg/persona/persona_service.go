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

// ListDialoguesByPersonaID returns dialogues for one stored persona row.
func (s *Service) ListDialoguesByPersonaID(personaID int64) ([]PersonaDialogueView, error) {
	rows, err := s.store.ListDialoguesByPersonaID(context.Background(), personaID)
	if err != nil {
		s.logger.Error("failed to list persona dialogues", slog.Int64("persona_id", personaID), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to list persona dialogues: %w", err)
	}
	return rows, nil
}
