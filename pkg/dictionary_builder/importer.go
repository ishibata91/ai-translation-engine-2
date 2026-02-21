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
}

// NewImporter creates a new instance of DictionaryImporter.
func NewImporter(config Config, store DictionaryStore) DictionaryImporter {
	return &xmlImporter{
		config: config,
		store:  store,
	}
}

// ImportXML reads an xTranslator XML from io.Reader using a streaming parser
// to extract allowed noun records and persist them.
func (i *xmlImporter) ImportXML(ctx context.Context, file io.Reader) (int, error) {
	slog.DebugContext(ctx, "ENTER DictionaryImporter.ImportXML")
	defer slog.DebugContext(ctx, "EXIT DictionaryImporter.ImportXML")

	decoder := xml.NewDecoder(file)

	// State to keep track of globally scoped values
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

		switch se := t.(type) {
		case xml.StartElement:
			// Check for addon name
			if se.Name.Local == "Addon" {
				var addon string
				err = decoder.DecodeElement(&addon, &se)
				if err == nil && addonName == "" {
					addonName = addon
				}
			} else if se.Name.Local == "String" {
				var strElem struct {
					EDID   string `xml:"EDID"`
					REC    string `xml:"REC"`
					Source string `xml:"Source"`
					Dest   string `xml:"Dest"`
				}

				err = decoder.DecodeElement(&strElem, &se)
				if err != nil {
					slog.WarnContext(ctx, "failed to decode String element, skipping", "error", err)
					continue
				}

				// Check if the REC type is allowed
				if !i.config.IsAllowedREC(strElem.REC) {
					continue
				}

				term := DictTerm{
					EDID:   strElem.EDID,
					REC:    strElem.REC,
					Source: strElem.Source,
					Dest:   strElem.Dest,
					Addon:  addonName,
				}

				batch = append(batch, term)

				if len(batch) >= batchSize {
					if err := i.store.SaveTerms(ctx, batch); err != nil {
						return totalImported, fmt.Errorf("error saving batch: %w", err)
					}
					totalImported += len(batch)
					batch = batch[:0]
				}
			}
		}
	}

	// Upsert any remaining entries
	if len(batch) > 0 {
		if err := i.store.SaveTerms(ctx, batch); err != nil {
			return totalImported, fmt.Errorf("error saving final batch: %w", err)
		}
		totalImported += len(batch)
	}

	slog.InfoContext(ctx, "Successfully imported terms", "total", totalImported)
	return totalImported, nil
}
