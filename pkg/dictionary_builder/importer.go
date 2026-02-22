package dictionary_builder

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
)

type xmlImporter struct {
	config Config
	store  DictionaryStore
	logger *slog.Logger
}

// NewImporter creates a new instance of DictionaryImporter.
func NewImporter(config Config, store DictionaryStore, logger *slog.Logger) DictionaryImporter {
	return &xmlImporter{
		config: config,
		store:  store,
		logger: logger.With("component", "DictionaryImporter"),
	}
}

// ImportXML reads an xTranslator XML from io.Reader using a streaming parser
// to extract allowed noun records and persist them.
func (i *xmlImporter) ImportXML(ctx context.Context, file io.Reader) (int, error) {
	i.logger.DebugContext(ctx, "ENTER DictionaryImporter.ImportXML")
	defer i.logger.DebugContext(ctx, "EXIT DictionaryImporter.ImportXML")

	decoder := xml.NewDecoder(file)

	var addonName string
	const batchSize = 1000
	batch := make([]DictTerm, 0, batchSize)
	totalImported := 0

	for {
		t, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return totalImported, fmt.Errorf("error reading xml token: %w", err)
		}

		if t == nil {
			break
		}

		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}

		switch se.Name.Local {
		case "Addon":
			addonName = i.handleAddonElement(ctx, decoder, &se, addonName)
		case "String":
			term, ok := i.handleStringElement(ctx, decoder, &se, addonName)
			if !ok {
				continue
			}
			batch = append(batch, term)

			if len(batch) >= batchSize {
				flushed, err := i.flushBatch(ctx, batch)
				if err != nil {
					return totalImported, err
				}
				totalImported += flushed
				batch = batch[:0]
			}
		}
	}

	// Flush remaining entries
	if len(batch) > 0 {
		flushed, err := i.flushBatch(ctx, batch)
		if err != nil {
			return totalImported, err
		}
		totalImported += flushed
	}

	i.logger.InfoContext(ctx, "Successfully imported terms", "total", totalImported)
	return totalImported, nil
}

// handleAddonElement decodes an Addon element and returns the addon name.
func (i *xmlImporter) handleAddonElement(ctx context.Context, decoder *xml.Decoder, se *xml.StartElement, currentAddon string) string {
	i.logger.DebugContext(ctx, "ENTER DictionaryImporter.handleAddonElement")

	var addon string
	err := decoder.DecodeElement(&addon, se)
	if err == nil && currentAddon == "" {
		return addon
	}
	return currentAddon
}

// handleStringElement decodes a String element and returns a DictTerm if the REC type is allowed.
func (i *xmlImporter) handleStringElement(ctx context.Context, decoder *xml.Decoder, se *xml.StartElement, addonName string) (DictTerm, bool) {
	i.logger.DebugContext(ctx, "ENTER DictionaryImporter.handleStringElement")

	var strElem struct {
		EDID   string `xml:"EDID"`
		REC    string `xml:"REC"`
		Source string `xml:"Source"`
		Dest   string `xml:"Dest"`
	}

	err := decoder.DecodeElement(&strElem, se)
	if err != nil {
		i.logger.WarnContext(ctx, "failed to decode String element, skipping", "error", err)
		return DictTerm{}, false
	}

	if !i.config.IsAllowedREC(strElem.REC) {
		return DictTerm{}, false
	}

	return DictTerm{
		EDID:   strElem.EDID,
		REC:    strElem.REC,
		Source: strElem.Source,
		Dest:   strElem.Dest,
		Addon:  addonName,
	}, true
}

// flushBatch persists a batch of terms to the store and returns the count saved.
func (i *xmlImporter) flushBatch(ctx context.Context, batch []DictTerm) (int, error) {
	i.logger.DebugContext(ctx, "ENTER DictionaryImporter.flushBatch", slog.Int("batchSize", len(batch)))

	if err := i.store.SaveTerms(ctx, batch); err != nil {
		return 0, fmt.Errorf("error saving batch: %w", err)
	}
	return len(batch), nil
}
