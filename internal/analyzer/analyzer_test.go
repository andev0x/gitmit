package analyzer

import (
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
