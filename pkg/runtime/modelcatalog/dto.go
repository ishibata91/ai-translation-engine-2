package modelcatalog

// ListModelsInput is a UI-facing request for fetching selectable models.
// Namespace is used to resolve provider/endpoint/api_key defaults from config.
type ListModelsInput struct {
	Namespace string `json:"namespace"`
	Provider  string `json:"provider"`
	Endpoint  string `json:"endpoint"`
	APIKey    string `json:"apiKey"`
}

// ModelCapability is provider-derived model execution capability exposed to UI.
type ModelCapability struct {
	SupportsBatch bool `json:"supports_batch"`
}

// ModelOption is a UI-facing model entry.
type ModelOption struct {
	ID               string          `json:"id"`
	DisplayName      string          `json:"display_name"`
	MaxContextLength int             `json:"max_context_length,omitempty"`
	Loaded           bool            `json:"loaded"`
	Capability       ModelCapability `json:"capability"`
}
