package modelcatalog

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewModelCatalogService,
	wire.Bind(new(Service), new(*ModelCatalogService)),
)
