package errorwrapcheck_test

import (
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/tools/backendquality/errorwrapcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, errorwrapcheck.Analyzer, "errorwrapcases")
}
