package controller

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/modelcatalog"
)

// ModelCatalogController exposes Wails-facing model catalog operations.
type ModelCatalogController struct {
	ctx     context.Context
	service *modelcatalog.ModelCatalogService
}

// NewModelCatalogController constructs the model catalog controller adapter.
func NewModelCatalogController(service *modelcatalog.ModelCatalogService) *ModelCatalogController {
	return &ModelCatalogController{
		ctx:     context.Background(),
		service: service,
	}
}

// SetContext injects the Wails application context for downstream propagation.
func (c *ModelCatalogController) SetContext(ctx context.Context) {
	if ctx == nil {
		c.ctx = context.Background()
		return
	}
	c.ctx = ctx
}

// ListModels returns selectable models for the given UI input.
func (c *ModelCatalogController) ListModels(input modelcatalog.ListModelsInput) ([]modelcatalog.ModelOption, error) {
	return c.service.ListModels(input)
}
