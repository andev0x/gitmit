package analyzer

import (
	"github.com/andev0x/gitmit/internal/config"
	"github.com/andev0x/gitmit/internal/parser"
	"testing"
)

func TestNormalizedScoring(t *testing.T) {
	cfg := &config.Config{
		NormalizeScoring: true,
		SignalWeights: map[string]float64{
			"branch":   0.35,
			"diffStat": 0.25,
			"keywords": 0.25,
			"patterns": 0.15,
		},
		Keywords: map[string]map[string]int{
			"fix": {"error": 4},
		},
	}

	t.Run("Branch signal dominates keyword", func(t *testing.T) {
		a := &Analyzer{
			config: cfg,
			changes: []*parser.Change{
				{File: "main.go", Action: "M", Diff: "+ var x = \"error\""},
			},
		}
		// branch "feat/new-ui" -> feat: 0.35 * 1.0 = 0.35
		// keyword "error" -> fix: 0.25 * 1.0 = 0.25
		// feat should win
		msg := a.AnalyzeChanges(1, 0, "feat/new-ui")
		if msg.Action != "feat" {
			t.Errorf("Expected action feat, got %s", msg.Action)
		}
	})

	t.Run("Keywords dominate if branch is missing", func(t *testing.T) {
		a := &Analyzer{
			config: cfg,
			changes: []*parser.Change{
				{File: "main.go", Action: "M", Diff: "+ var x = \"error\""},
			},
		}
		// keyword "error" -> fix: 0.25 * 1.0 = 0.25
		// 0.25 < 0.35 (fallback threshold)
		// So it should fallback to determineAction which for Action: "M" is refactor
		msg := a.AnalyzeChanges(1, 0, "")
		if msg.Action != "refactor" {
			t.Errorf("Expected action refactor (fallback), got %s", msg.Action)
		}
	})

	t.Run("Combined signals work together", func(t *testing.T) {
		a := &Analyzer{
			config: cfg,
			changes: []*parser.Change{
				{File: "main.go", Action: "M", Diff: "+ func NewFeature() {", Added: 40, Removed: 0},
			},
		}
		// branch "feature/cool" -> feat: 0.35
		// ratio 1.0 -> feat: 0.25 * 1.0 = 0.25
		// total feat = 0.60
		msg := a.AnalyzeChanges(40, 0, "feature/cool")
		if msg.Action != "feat" {
			t.Errorf("Expected action feat, got %s", msg.Action)
		}
	})

	t.Run("Fallback to additive if disabled", func(t *testing.T) {
		cfgDisabled := *cfg
		cfgDisabled.NormalizeScoring = false
		a := &Analyzer{
			config: &cfgDisabled,
			changes: []*parser.Change{
				{File: "main.go", Action: "M", Diff: "+ var x = \"error\""},
			},
		}
		// In additive:
		// branch "feat/new-ui" -> feat: 3
		// keyword "error" -> fix: 4
		// fix should win
		msg := a.AnalyzeChanges(1, 0, "feat/new-ui")
		if msg.Action != "fix" {
			t.Errorf("Expected action fix, got %s", msg.Action)
		}
	})
}
