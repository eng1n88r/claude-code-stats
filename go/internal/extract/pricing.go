package extract

// ModelPricing holds per-million-token rates for a model.
type ModelPricing struct {
	Input        float64
	Output       float64
	CacheRead    float64
	CacheWrite5m float64
	CacheWrite1h float64
	Display      string
}

var pricing = map[string]ModelPricing{
	"claude-opus-4-6": {
		Input: 5.00, Output: 25.00,
		CacheRead: 0.50, CacheWrite5m: 6.25, CacheWrite1h: 10.00,
		Display: "Opus 4.6",
	},
	"claude-opus-4-5-20251101": {
		Input: 5.00, Output: 25.00,
		CacheRead: 0.50, CacheWrite5m: 6.25, CacheWrite1h: 10.00,
		Display: "Opus 4.5",
	},
	"claude-sonnet-4-5-20250929": {
		Input: 3.00, Output: 15.00,
		CacheRead: 0.30, CacheWrite5m: 3.75, CacheWrite1h: 6.00,
		Display: "Sonnet 4.5",
	},
	"claude-haiku-4-5-20251001": {
		Input: 1.00, Output: 5.00,
		CacheRead: 0.10, CacheWrite5m: 1.25, CacheWrite1h: 2.00,
		Display: "Haiku 4.5",
	},
}

var defaultPricing = ModelPricing{
	Input: 5.00, Output: 25.00,
	CacheRead: 0.50, CacheWrite5m: 6.25, CacheWrite1h: 10.00,
	Display: "Unknown",
}

// GetPricing returns the pricing for a model, falling back to default.
func GetPricing(modelID string) ModelPricing {
	if p, ok := pricing[modelID]; ok {
		return p
	}
	return defaultPricing
}

// GetModelDisplay returns the human-readable display name for a model.
func GetModelDisplay(modelID string) string {
	return GetPricing(modelID).Display
}

// CalcCost calculates the cost for a single API call based on token counts.
// Uses the standard cache write rate (cache_write_5m) for all cache creation
// tokens, matching Claude Code's own cost calculation.
func CalcCost(modelID string, inputTokens, outputTokens, cacheRead, cacheCreation int) float64 {
	p := GetPricing(modelID)
	return float64(inputTokens)*p.Input/1_000_000 +
		float64(outputTokens)*p.Output/1_000_000 +
		float64(cacheRead)*p.CacheRead/1_000_000 +
		float64(cacheCreation)*p.CacheWrite5m/1_000_000
}

// ProjectDisplayName extracts a short display name from a project path.
func ProjectDisplayName(projectPath string) string {
	if projectPath == "" {
		return "Unknown"
	}
	// Normalize backslashes
	p := projectPath
	for i := range p {
		if p[i] == '\\' {
			p = p[:i] + "/" + p[i+1:]
		}
	}
	// Trim trailing slash
	for len(p) > 0 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	if p == "" {
		return projectPath
	}
	// Split and take last 2 parts
	parts := splitPath(p)
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return projectPath
}

func splitPath(p string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(p); i++ {
		if p[i] == '/' {
			if i > start {
				parts = append(parts, p[start:i])
			}
			start = i + 1
		}
	}
	if start < len(p) {
		parts = append(parts, p[start:])
	}
	return parts
}
