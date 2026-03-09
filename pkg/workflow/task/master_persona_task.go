package task

import (
	"context"
	"fmt"
)

type StartMasterPersonTaskInput struct {
	SourceJSONPath    string `json:"source_json_path"`
	OverwriteExisting bool   `json:"overwrite_existing"`
}

func (b *Bridge) StartMasterPersonTask(ctx context.Context, input StartMasterPersonTaskInput) (string, error) {
	if b.masterPersonaWorkflow == nil {
		return "", fmt.Errorf("master persona workflow is not configured")
	}
	return b.masterPersonaWorkflow.StartMasterPersonTask(ctx, input)
}
