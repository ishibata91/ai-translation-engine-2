package shared

import "context"

// SliceArtifactRef identifies one shared artifact unit produced by one slice.
type SliceArtifactRef struct {
	SliceID string `json:"slice_id"`
	RefID   string `json:"ref_id"`
}

// SearchQuery defines common lookup conditions for shared artifacts.
type SearchQuery struct {
	SliceID    string                 `json:"slice_id"`
	RefID      string                 `json:"ref_id,omitempty"`
	Keywords   []string               `json:"keywords,omitempty"`
	Cursor     string                 `json:"cursor,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// Record represents one shared artifact search hit.
type Record struct {
	RefID      string                 `json:"ref_id"`
	SliceID    string                 `json:"slice_id"`
	Payload    map[string]interface{} `json:"payload"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// SearchStore provides shared artifact search capability across slices.
type SearchStore interface {
	Find(ctx context.Context, query SearchQuery) ([]Record, string, error)
}
