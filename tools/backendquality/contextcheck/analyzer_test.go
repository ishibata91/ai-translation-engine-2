package contextcheck_test

import (
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/tools/backendquality/contextcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, contextcheck.Analyzer, "contextcases")
}
