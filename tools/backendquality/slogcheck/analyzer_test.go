package slogcheck_test

import (
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/tools/backendquality/slogcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, slogcheck.Analyzer, "slogcases")
}
