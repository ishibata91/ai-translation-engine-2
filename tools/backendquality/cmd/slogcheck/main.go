package main

import (
	"github.com/ishibata91/ai-translation-engine-2/tools/backendquality/slogcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(slogcheck.Analyzer)
}
