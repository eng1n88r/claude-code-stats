package extract

import (
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestParityWithPython runs the Python extract on a shared JSONL fixture
// and compares key outputs with the Go implementation.
// Requires Python 3 available as "python" in PATH.
func TestParityWithPython(t *testing.T) {
	pythonScript := filepath.Join("..", "..", "..", "python", "test_parity_bridge.py")
	fixturePath := filepath.Join("..", "..", "..", "testdata", "sample_session.jsonl")

	if _, err := os.Stat(pythonScript); os.IsNotExist(err) {
		t.Skip("python/test_parity_bridge.py not found, skipping parity test")
	}
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("testdata/sample_session.jsonl not found, skipping parity test")
	}

	// Run Python bridge script which parses the same fixture
	cmd := exec.Command("python", pythonScript, fixturePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Python bridge failed: %v\nOutput: %s", err, out)
	}

	var pyResult struct {
		SessionID       string         `json:"session_id"`
		Messages        int            `json:"messages"`
		UserMessages    int            `json:"user_messages"`
		AssistMessages  int            `json:"assistant_messages"`
		FirstPrompt     string         `json:"first_prompt"`
		Models          map[string]int `json:"model_calls"`
		TotalCost       float64        `json:"total_cost"`
		TotalInput      int            `json:"total_input"`
		TotalOutput     int            `json:"total_output"`
		TotalCacheRead  int            `json:"total_cache_read"`
		TotalCacheWrite int            `json:"total_cache_write"`
		Tools           map[string]int `json:"tools"`
		ProjectPath     string         `json:"project_path"`
	}
	if err := json.Unmarshal(out, &pyResult); err != nil {
		t.Fatalf("Failed to parse Python output: %v\nRaw: %s", err, out)
	}

	// Now parse the same fixture with Go
	sessions := make(map[string]*rawSession)
	parseJSONLFile(fixturePath, "fallback-id", "test-project", 1024, "test", sessions)

	sess, ok := sessions["test-session-001"]
	if !ok {
		t.Fatal("Go: session test-session-001 not found")
	}

	// Compare counts
	if sess.MessageCount != pyResult.Messages {
		t.Errorf("Messages: Go=%d, Python=%d", sess.MessageCount, pyResult.Messages)
	}
	if sess.UserMsgCount != pyResult.UserMessages {
		t.Errorf("UserMessages: Go=%d, Python=%d", sess.UserMsgCount, pyResult.UserMessages)
	}
	if sess.AssistMsgCount != pyResult.AssistMessages {
		t.Errorf("AssistMessages: Go=%d, Python=%d", sess.AssistMsgCount, pyResult.AssistMessages)
	}
	if sess.FirstPrompt != pyResult.FirstPrompt {
		t.Errorf("FirstPrompt: Go=%q, Python=%q", sess.FirstPrompt, pyResult.FirstPrompt)
	}
	if sess.ProjectPath != pyResult.ProjectPath {
		t.Errorf("ProjectPath: Go=%q, Python=%q", sess.ProjectPath, pyResult.ProjectPath)
	}

	// Compare model call counts
	for model, goAccum := range sess.Models {
		pyCalls, ok := pyResult.Models[model]
		if !ok {
			t.Errorf("Model %s exists in Go but not Python", model)
			continue
		}
		if goAccum.Calls != pyCalls {
			t.Errorf("Model %s calls: Go=%d, Python=%d", model, goAccum.Calls, pyCalls)
		}
	}

	// Compare total tokens
	goInput, goOutput, goCacheRead, goCacheWrite := 0, 0, 0, 0
	goCost := 0.0
	for _, m := range sess.Models {
		goInput += m.InputTokens
		goOutput += m.OutputTokens
		goCacheRead += m.CacheRead
		goCacheWrite += m.CacheCreation
		goCost += m.Cost
	}

	if goInput != pyResult.TotalInput {
		t.Errorf("TotalInput: Go=%d, Python=%d", goInput, pyResult.TotalInput)
	}
	if goOutput != pyResult.TotalOutput {
		t.Errorf("TotalOutput: Go=%d, Python=%d", goOutput, pyResult.TotalOutput)
	}
	if goCacheRead != pyResult.TotalCacheRead {
		t.Errorf("TotalCacheRead: Go=%d, Python=%d", goCacheRead, pyResult.TotalCacheRead)
	}
	if goCacheWrite != pyResult.TotalCacheWrite {
		t.Errorf("TotalCacheWrite: Go=%d, Python=%d", goCacheWrite, pyResult.TotalCacheWrite)
	}
	if math.Abs(goCost-pyResult.TotalCost) > 0.000001 {
		t.Errorf("TotalCost: Go=%f, Python=%f", goCost, pyResult.TotalCost)
	}

	// Compare tool usage
	for tool, goCount := range sess.Tools {
		pyCount, ok := pyResult.Tools[tool]
		if !ok {
			t.Errorf("Tool %s exists in Go but not Python", tool)
			continue
		}
		if goCount != pyCount {
			t.Errorf("Tool %s: Go=%d, Python=%d", tool, goCount, pyCount)
		}
	}
}
