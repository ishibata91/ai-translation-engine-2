package controller

import (
	"context"
	"fmt"

	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/persona"
)

// PersonaController exposes Wails-facing persona read operations.
type PersonaController struct {
	ctx     context.Context
	service *persona.Service
}

// NewPersonaController constructs the persona controller adapter.
func NewPersonaController(service *persona.Service) *PersonaController {
	return &PersonaController{
		ctx:     context.Background(),
		service: service,
	}
}

// SetContext injects the Wails application context for downstream propagation.
func (c *PersonaController) SetContext(ctx context.Context) {
	if ctx == nil {
		c.ctx = context.Background()
		return
	}
	c.ctx = ctx
}

// ListNPCs returns persona rows for UI table rendering.
func (c *PersonaController) ListNPCs() ([]persona.PersonaNPCView, error) {
	return c.service.ListNPCs()
}

// ListDialoguesByPersonaID returns dialogues for one stored persona.
func (c *PersonaController) ListDialoguesByPersonaID(personaID int64) ([]persona.PersonaDialogueView, error) {
	if personaID <= 0 {
		return nil, fmt.Errorf("persona_id is required")
	}
	return c.service.ListDialoguesByPersonaID(personaID)
}
