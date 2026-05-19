package analyzer

import (
	"github.com/andev0x/gitmit/internal/config"
	"github.com/andev0x/gitmit/internal/parser"
	"testing"
)

func TestParseBranchName(t *testing.T) {
	a := &Analyzer{}

	tests := []struct {
		branch   string
		wantType string
		wantScope string
	}{
		{"feature/auth-login", "feat", "auth"},
		{"feat/ui-button", "feat", "ui"},
		{"bugfix/fix-memleak", "fix", "fix"},
		{"fix/typo", "fix", "typo"},
		{"hotfix/urgent-patch", "fix", "urgent"},
		{"refactor/api-cleanup", "refactor", "api"},
		{"chore/deps-update", "chore", "deps"},
		{"docs/readme-update", "docs", "readme"},
		{"feature/login", "feat", "login"},
		{"random-branch", "", ""},
	}

	for _, tt := range tests {
		gotType, gotScope := a.parseBranchName(tt.branch)
		if gotType != tt.wantType {
			t.Errorf("parseBranchName(%q) gotType = %q, want %q", tt.branch, gotType, tt.wantType)
		}
		if gotScope != tt.wantScope {
			t.Errorf("parseBranchName(%q) gotScope = %q, want %q", tt.branch, gotScope, tt.wantScope)
		}
	}
}

func TestCalculateHistoryScope(t *testing.T) {
	a := &Analyzer{}

	tests := []struct {
		name    string
		commits []string
		want    string
	}{
		{
			"More than 50% frequency",
			[]string{
				"feat(auth): login",
				"fix(auth): redirect",
				"feat(auth): signup",
				"chore: update deps",
				"docs: update readme",
			},
			"auth",
		},
		{
			"Exactly 50% frequency",
			[]string{
				"feat(auth): login",
				"fix(auth): redirect",
				"feat(ui): button",
				"chore: update deps",
			},
			"", // 2/4 is not > 50%
		},
		{
			"No scopes",
			[]string{
				"feat: login",
				"fix: redirect",
			},
			"",
		},
		{
			"Empty list",
			[]string{},
			"",
		},
		{
			"Different scopes",
			[]string{
				"feat(auth): login",
				"feat(ui): button",
				"feat(db): query",
				"feat(api): endpoint",
				"feat(docs): page",
			},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a.calculateHistoryScope(tt.commits)
			if got != tt.want {
				t.Errorf("calculateHistoryScope() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectNewDependencies(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		diff     string
		want     []string
	}{
		{
			"Go mod addition",
			"go.mod",
			"+	github.com/stretchr/testify v1.8.0\n+	github.com/spf13/cobra v1.5.0",
			[]string{"github.com/stretchr/testify", "github.com/spf13/cobra"},
		},
		{
			"Package JSON addition",
			"package.json",
			"+    \"lodash\": \"^4.17.21\",\n+    \"react\": \"^18.2.0\"",
			[]string{"lodash", "react"},
		},
		{
			"Requirements TXT addition",
			"requirements.txt",
			"+requests==2.28.1\n+flask==2.2.2",
			[]string{"requests", "flask"},
		},
		{
			"Cargo TOML addition",
			"Cargo.toml",
			"+serde = \"1.0\"\n+tokio = \"1.0\"",
			[]string{"serde", "tokio"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Analyzer{
				changes: []*parser.Change{
					{File: tt.fileName, Diff: tt.diff},
				},
			}
			got := a.detectNewDependencies()
			if len(got) != len(tt.want) {
				t.Errorf("detectNewDependencies() got = %v, want %v", got, tt.want)
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("detectNewDependencies()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestAnalyzeDiffStatRatio(t *testing.T) {
	a := &Analyzer{config: &config.Config{}}

	tests := []struct {
		added   int
		removed int
		want    string
	}{
		{10, 90, "refactor"}, // Ratio 0.1 < 0.2
		{90, 10, "feat"},     // Ratio 0.9 > 0.8 and added > 30
		{50, 50, "refactor"}, // Ratio 0.5 balanced
		{10, 10, "refactor"}, // Ratio 0.5 balanced
		{20, 2, "feat"},      // Ratio 0.9 > 0.8 but added < 30 -> empty from this func, defaults elsewhere
	}

	for _, tt := range tests {
		got := a.analyzeDiffStat(tt.added, tt.removed)
		if tt.want == "feat" && tt.added < 30 {
			if got != "" {
				t.Errorf("analyzeDiffStat(%d, %d) = %q, want \"\"", tt.added, tt.removed, got)
			}
		} else if got != tt.want {
			t.Errorf("analyzeDiffStat(%d, %d) = %q, want %q", tt.added, tt.removed, got, tt.want)
		}
	}
}

func TestStructureDetectionRegex(t *testing.T) {
	a := &Analyzer{}

	t.Run("Go functions and structs", func(t *testing.T) {
		diff := "+func MyFunc() {\n+type MyStruct struct {\n+func (r *Receiver) MyMethod() {"
		funcs := a.detectFunctions(diff)
		structs := a.detectStructs(diff)

		if !contains(funcs, "MyFunc") {
			t.Errorf("Expected MyFunc in %v", funcs)
		}
		if !contains(structs, "MyStruct") {
			t.Errorf("Expected MyStruct in %v", structs)
		}
	})

	t.Run("TS functions and classes", func(t *testing.T) {
		diff := "+function myFunc() {\n+const myArrow = () => {\n+class MyClass {"
		funcs := a.detectFunctions(diff)
		structs := a.detectStructs(diff)

		if !contains(funcs, "myFunc") {
			t.Errorf("Expected myFunc in %v", funcs)
		}
		if !contains(funcs, "myArrow") {
			t.Errorf("Expected myArrow in %v", funcs)
		}
		if !contains(structs, "MyClass") {
			t.Errorf("Expected MyClass in %v", structs)
		}
	})

	t.Run("Python functions and classes", func(t *testing.T) {
		diff := "+def my_func():\n+class MyClass:"
		funcs := a.detectFunctions(diff)
		structs := a.detectStructs(diff)

		if !contains(funcs, "my_func") {
			t.Errorf("Expected my_func in %v", funcs)
		}
		if !contains(structs, "MyClass") {
			t.Errorf("Expected MyClass in %v", structs)
		}
	})
}

func TestCrossScoringMatrix(t *testing.T) {
	cfg := &config.Config{
		Keywords: map[string]map[string]int{
			"fix": {"error": 4},
		},
	}

	t.Run("Branch overrides keyword if score is higher", func(t *testing.T) {
		a := &Analyzer{
			config: cfg,
			changes: []*parser.Change{
				{File: "main.go", Diff: "+ fmt.Println(\"error\")"},
			},
		}
		// branch "feat/new-ui" adds 3 to feat
		// "error" keyword adds 4 to fix
		// fix (4) > feat (3) -> fix
		msg := a.AnalyzeChanges(1, 0, "feat/new-ui")
		if msg.Action != "fix" {
			t.Errorf("Expected action fix, got %s", msg.Action)
		}
	})

	t.Run("Combined signals", func(t *testing.T) {
		a := &Analyzer{
			config: cfg,
			changes: []*parser.Change{
				{File: "main.go", Diff: "+ func NewFeature() {", Added: 40, Removed: 0},
			},
		}
		// branch "feature/cool" adds 3 to feat
		// ratio 1.0 adds 2 to feat (added > 30)
		// total feat = 5
		msg := a.AnalyzeChanges(40, 0, "feature/cool")
		if msg.Action != "feat" {
			t.Errorf("Expected action feat, got %s", msg.Action)
		}
	})
}
