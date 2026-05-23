package formatter

import (
	"testing"
)

func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name             string
		msg              string
		maxSubject       int
		maxBody          int
		expected         string
	}{
		{
			name:       "short subject, no wrapping",
			msg:        "feat: add feature",
			maxSubject: 50,
			maxBody:    72,
			expected:   "feat: add feature",
		},
		{
			name:       "long subject, wrap at 10",
			msg:        "feat: add new feature for login",
			maxSubject: 10,
			maxBody:    72,
			expected:   "feat: add\n\nnew feature for login",
		},
		{
			name:       "subject and body, no wrapping",
			msg:        "feat: add feature\n\nThis is a body message.",
			maxSubject: 50,
			maxBody:    72,
			expected:   "feat: add feature\n\nThis is a body message.",
		},
		{
			name:       "subject and body, wrap both",
			msg:        "feat: add feature\n\nThis is a body message that is very long.",
			maxSubject: 10,
			maxBody:    10,
			expected:   "feat: add\n\nfeature\n\nThis is a\nbody\nmessage\nthat is\nvery long.",
		},
		{
			name:       "redundant phrases",
			msg:        "feat feat: add add new feature",
			maxSubject: 50,
			maxBody:    72,
			expected:   "feat: add new feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.maxSubject, tt.maxBody)
			actual := f.FormatMessage(tt.msg, false)
			if actual != tt.expected {
				t.Errorf("FormatMessage() = %q, want %q", actual, tt.expected)
			}
		})
	}
}
