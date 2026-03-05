package extract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int64
	}{
		{
			name:  "ISO 8601 string",
			input: "2026-02-10T14:30:00Z",
			want:  1770733800000,
		},
		{
			name:  "Unix milliseconds float64",
			input: float64(1738504200000),
			want:  1738504200000,
		},
		{
			name:  "invalid string",
			input: "not-a-date",
			want:  0,
		},
		{
			name:  "nil",
			input: nil,
			want:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTimestamp(tt.input)
			if got != tt.want {
				t.Errorf("parseTimestamp(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractPromptText(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  string
	}{
		{
			name: "string content",
			input: map[string]interface{}{
				"message": map[string]interface{}{
					"content": "Hello world",
				},
			},
			want: "Hello world",
		},
		{
			name: "array content with text block",
			input: map[string]interface{}{
				"message": map[string]interface{}{
					"content": []interface{}{
						map[string]interface{}{"type": "text", "text": "Fix the bug"},
					},
				},
			},
			want: "Fix the bug",
		},
		{
			name:  "no message",
			input: map[string]interface{}{},
			want:  "",
		},
		{
			name: "empty content array",
			input: map[string]interface{}{
				"message": map[string]interface{}{
					"content": []interface{}{},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPromptText(tt.input)
			if got != tt.want {
				t.Errorf("extractPromptText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseJSONLFile(t *testing.T) {
	// Use the shared test fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "sample_session.jsonl")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("testdata/sample_session.jsonl not found, skipping")
	}

	sessions := make(map[string]*rawSession)
	lines := parseJSONLFile(fixturePath, "fallback-id", "test-project", 1024, "test", sessions)

	if lines != 6 {
		t.Errorf("expected 6 lines parsed, got %d", lines)
	}

	sess, ok := sessions["test-session-001"]
	if !ok {
		t.Fatal("session test-session-001 not found")
	}

	// Check message counts
	if sess.MessageCount != 6 {
		t.Errorf("expected 6 messages, got %d", sess.MessageCount)
	}
	if sess.UserMsgCount != 3 {
		t.Errorf("expected 3 user messages, got %d", sess.UserMsgCount)
	}
	if sess.AssistMsgCount != 3 {
		t.Errorf("expected 3 assistant messages, got %d", sess.AssistMsgCount)
	}

	// Check first prompt extraction (should skip <command and [Request interrupted)
	if sess.FirstPrompt != "Hello, help me fix the login bug" {
		t.Errorf("unexpected first prompt: %q", sess.FirstPrompt)
	}

	// Check model accumulation — should have 2 models
	if len(sess.Models) != 2 {
		t.Errorf("expected 2 models, got %d", len(sess.Models))
	}

	opus := sess.Models["claude-opus-4-5-20251101"]
	if opus == nil {
		t.Fatal("opus model accumulator not found")
	}
	if opus.Calls != 2 {
		t.Errorf("expected 2 opus calls, got %d", opus.Calls)
	}
	if opus.InputTokens != 3000 {
		t.Errorf("expected 3000 input tokens, got %d", opus.InputTokens)
	}
	if opus.OutputTokens != 1300 {
		t.Errorf("expected 1300 output tokens, got %d", opus.OutputTokens)
	}

	haiku := sess.Models["claude-haiku-4-5-20251001"]
	if haiku == nil {
		t.Fatal("haiku model accumulator not found")
	}
	if haiku.Calls != 1 {
		t.Errorf("expected 1 haiku call, got %d", haiku.Calls)
	}

	// Check tool usage
	if sess.Tools["Read"] != 1 {
		t.Errorf("expected 1 Read tool use, got %d", sess.Tools["Read"])
	}
	if sess.Tools["Edit"] != 1 {
		t.Errorf("expected 1 Edit tool use, got %d", sess.Tools["Edit"])
	}

	// Check project path
	if sess.ProjectPath != "/home/user/projects/myapp" {
		t.Errorf("unexpected project path: %q", sess.ProjectPath)
	}
}

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	if got := getString(m, "key1"); got != "value1" {
		t.Errorf("getString for key1 = %q, want 'value1'", got)
	}
	if got := getString(m, "key2"); got != "" {
		t.Errorf("getString for key2 = %q, want ''", got)
	}
	if got := getString(m, "missing"); got != "" {
		t.Errorf("getString for missing = %q, want ''", got)
	}
}

func TestGetInt(t *testing.T) {
	m := map[string]interface{}{
		"float_key":  float64(42),
		"string_key": "not-a-number",
	}
	if got := getInt(m, "float_key"); got != 42 {
		t.Errorf("getInt for float_key = %d, want 42", got)
	}
	if got := getInt(m, "string_key"); got != 0 {
		t.Errorf("getInt for string_key = %d, want 0", got)
	}
	if got := getInt(m, "missing"); got != 0 {
		t.Errorf("getInt for missing = %d, want 0", got)
	}
}
