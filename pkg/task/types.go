package task

import (
	"encoding/json"
	"time"
)

type TaskStatus string

const (
	StatusRunning   TaskStatus = "running"
	StatusPaused    TaskStatus = "paused"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusCancelled TaskStatus = "cancelled"
)

type TaskType string

const (
	TypeDictionaryBuild    TaskType = "dictionary_build"
	TypePersonaExtraction  TaskType = "persona_extraction"
	TypeTranslationProject TaskType = "translation_project"
)

type Task struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Type      TaskType     `json:"type"`
	Status    TaskStatus   `json:"status"`
	Phase     string       `json:"phase"`
	Progress  float64      `json:"progress"`
	ErrorMsg  string       `json:"error_msg"`
	Metadata  TaskMetadata `json:"metadata"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type TaskMetadata map[string]interface{}

func (m TaskMetadata) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]interface{}(m))
}

func (m *TaskMetadata) UnmarshalJSON(data []byte) error {
	var tmp map[string]interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*m = tmp
	return nil
}
