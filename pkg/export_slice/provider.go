package export_slice

import (
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for ExportSlice.
var ProviderSet = wire.NewSet(
	NewExporter,
)
