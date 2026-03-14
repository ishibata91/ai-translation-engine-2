package personacontroller

import (
	"fmt"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/testenv"
)

// Env bundles persona controller test dependencies.
type Env struct {
	Service *FakeService
	TestEnv *testenv.Env
}

// FakeService stubs persona read service behavior.
type FakeService struct {
	NPCs           []persona.PersonaNPCView
	NPCsErr        error
	LastDialogueID int64
	Dialogues      []persona.PersonaDialogueView
	DialoguesErr   error
}

func (s *FakeService) ListNPCs() ([]persona.PersonaNPCView, error) {
	return s.NPCs, s.NPCsErr
}

func (s *FakeService) ListDialoguesByPersonaID(personaID int64) ([]persona.PersonaDialogueView, error) {
	s.LastDialogueID = personaID
	return s.Dialogues, s.DialoguesErr
}

// Build creates persona controller dependencies on shared testenv.
func Build(t *testing.T, name string) *Env {
	t.Helper()
	base := testenv.NewFileSQLiteEnv(t, name)
	return &Env{Service: &FakeService{}, TestEnv: base}
}

// String returns a short summary useful in failures.
func (e *Env) String() string {
	if e == nil || e.TestEnv == nil {
		return "<nil personacontroller env>"
	}
	return fmt.Sprintf("db=%s trace_id=%s", e.TestEnv.DBPath, testenv.TraceIDValue(e.TestEnv.Ctx))
}
