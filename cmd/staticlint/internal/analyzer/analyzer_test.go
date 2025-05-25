package analyzer_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/nekr0z/muhame/cmd/staticlint/internal/analyzer"
)

func TestMyAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.ExitAnalyzer, "./...")
}
