package main

import (
	"slices"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation"
)

func TestNewTermRecordConfigUsesFoundationRECTypes(t *testing.T) {
	got := newTermRecordConfig()
	if got == nil {
		t.Fatalf("newTermRecordConfig returned nil")
	}

	if !slices.Equal(got.TargetRecordTypes, foundation.DictionaryImportRECTypes) {
		t.Fatalf("target rec types must match foundation constant: got=%v want=%v", got.TargetRecordTypes, foundation.DictionaryImportRECTypes)
	}

	if !slices.Contains(got.TargetRecordTypes, "CONT:FULL") {
		t.Fatalf("CONT:FULL must be included in target rec types: got=%v", got.TargetRecordTypes)
	}
}
