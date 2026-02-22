package process_manager

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // For development. Should be restricted later.
	},
}

// Handler handles HTTP and WebSocket requests for ProcessManager.
type Handler struct {
	manager *Manager
	hub     *progress.Hub
}

// NewHandler creates a new Handler.
func NewHandler(manager *Manager, hub *progress.Hub) *Handler {
	return &Handler{manager: manager, hub: hub}
}

// ServeWebSocket upgrades the connection and streams progress events.
func (h *Handler) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.manager.logger.ErrorContext(r.Context(), "Failed to upgrade websocket", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	clientChan := make(chan progress.ProgressEvent, 20)
	h.hub.Subscribe(clientChan)
	defer h.hub.Unsubscribe(clientChan)

	// Read loop (to handle client close or ping/pong)
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()

	// Write loop
	for event := range clientChan {
		if err := conn.WriteJSON(event); err != nil {
			h.manager.logger.ErrorContext(r.Context(), "Failed to write websocket message",
				slog.String("correlation_id", event.CorrelationID),
				slog.String("error", err.Error()))
			break
		}
	}
}

// StartProcessRequest is the payload to start a new translation process.
type StartProcessRequest struct {
	Slice     string `json:"slice"`
	InputFile string `json:"input_file"`
}

// StartProcessResponse is the response containing the new ProcessID.
type StartProcessResponse struct {
	ProcessID string `json:"process_id"`
}

// HandleStartProcess is the REST endpoint to initiate a process.
func (h *Handler) HandleStartProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// ExecuteSlice handles state saving, job submission and worker initiation.
	processID, err := h.manager.ExecuteSlice(r.Context(), req.Slice, nil, req.InputFile)
	if err != nil {
		h.manager.logger.ErrorContext(r.Context(), "Failed to execute slice",
			slog.String("slice", req.Slice),
			slog.String("input_file", req.InputFile),
			slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("failed to start process: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StartProcessResponse{ProcessID: processID})
}

// RegisterRoutes registers the handlers to the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ws/progress", h.ServeWebSocket)
	mux.HandleFunc("/api/process/start", h.HandleStartProcess)
}
