package main

import (
	"github.com/ishibata91/ai-translation-engine-2/tools/backendquality/contextcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(contextcheck.Analyzer)
}
