package extract

import "encoding/json"

// DashboardData is the top-level output matching dashboard_data.json schema.
type DashboardData struct {
	GeneratedAt   string                 `json:"generated_at"`
	Locale        json.RawMessage        `json:"locale"`
	Account       Account                `json:"account"`
	KPI           KPI                    `json:"kpi"`
	Plan          PlanAnalysis           `json:"plan"`
	DailyCosts    []map[string]any       `json:"daily_costs"`
	Cumulative    []CumulativeCost       `json:"cumulative_costs"`
	DailyMessages []DailyMessage         `json:"daily_messages"`
	Hourly        []HourlyDist           `json:"hourly_distribution"`
	Weekday       []WeekdayDist          `json:"weekday_distribution"`
	Models        []string               `json:"models"`
	ModelSummary  []ModelSummary         `json:"model_summary"`
	CostByType    map[string]float64     `json:"cost_by_token_type"`
	Projects      []ProjectSummary       `json:"projects"`
	Sessions      []SessionOutput        `json:"sessions"`
	ToolSummary   []ToolCount            `json:"tool_summary"`
	Insights      Insights               `json:"insights"`
}

type Account struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type KPI struct {
	TotalCost       float64 `json:"total_cost"`
	ActualPlanCost  float64 `json:"actual_plan_cost"`
	TotalSessions   int     `json:"total_sessions"`
	TotalMessages   int     `json:"total_messages"`
	TotalOutput     int     `json:"total_output_tokens"`
	TotalInput      int     `json:"total_input_tokens"`
	FirstSession    string  `json:"first_session"`
	LastSession     string  `json:"last_session"`
	TotalProjects   int     `json:"total_projects"`
}

type PlanAnalysis struct {
	Periods        []PlanPeriod    `json:"periods"`
	CurrentBilling *CurrentBilling `json:"current_billing"`
	TotalAPICost   float64         `json:"total_api_cost"`
	TotalPlanCost  float64         `json:"total_plan_cost"`
	TotalSavings   float64         `json:"total_savings"`
	OverallROI     float64         `json:"overall_roi"`
}

type PlanPeriod struct {
	Plan        string   `json:"plan"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	TotalDays   int      `json:"total_days"`
	DaysActive  int      `json:"days_active"`
	PlanCostEUR *float64 `json:"plan_cost_eur"`
	PlanCostUSD float64  `json:"plan_cost_usd"`
	APICost     float64  `json:"api_cost"`
	Savings     float64  `json:"savings"`
	ROIFactor   float64  `json:"roi_factor"`
	Sessions    int      `json:"sessions"`
	Messages    int      `json:"messages"`
	CostPerDay  float64  `json:"cost_per_day"`
}

type CurrentBilling struct {
	Plan          string   `json:"plan"`
	PeriodStart   string   `json:"period_start"`
	PeriodEnd     string   `json:"period_end"`
	DaysElapsed   int      `json:"days_elapsed"`
	DaysTotal     int      `json:"days_total"`
	DaysRemaining int      `json:"days_remaining"`
	PlanCostEUR   *float64 `json:"plan_cost_eur"`
	PlanCostUSD   float64  `json:"plan_cost_usd"`
	APICost       float64  `json:"api_cost"`
	ProjectedCost float64  `json:"projected_cost"`
	Savings       float64  `json:"savings"`
	ROIFactor     float64  `json:"roi_factor"`
	Sessions      int      `json:"sessions"`
	Messages      int      `json:"messages"`
	CostPerDay    float64  `json:"cost_per_day"`
}

type CumulativeCost struct {
	Date string  `json:"date"`
	Cost float64 `json:"cost"`
}

type DailyMessage struct {
	Date     string `json:"date"`
	Messages int    `json:"messages"`
	Sessions int    `json:"sessions"`
}

type HourlyDist struct {
	Hour     int `json:"hour"`
	Messages int `json:"messages"`
}

type WeekdayDist struct {
	Day      string `json:"day"`
	Messages int    `json:"messages"`
}

type ModelSummary struct {
	Model           string  `json:"model"`
	Cost            float64 `json:"cost"`
	InputTokens     int     `json:"input_tokens"`
	OutputTokens    int     `json:"output_tokens"`
	CacheReadTokens int     `json:"cache_read_tokens"`
	CacheWriteTokens int    `json:"cache_write_tokens"`
	Calls           int     `json:"calls"`
}

type ProjectSummary struct {
	Name             string  `json:"name"`
	Sessions         int     `json:"sessions"`
	Messages         int     `json:"messages"`
	Cost             float64 `json:"cost"`
	InputTokens      int     `json:"input_tokens"`
	OutputTokens     int     `json:"output_tokens"`
	CacheReadTokens  int     `json:"cache_read_tokens"`
	CacheWriteTokens int     `json:"cache_write_tokens"`
	FileSizeMB       float64 `json:"file_size_mb"`
}

type SessionOutput struct {
	SessionID      string                    `json:"session_id"`
	Project        string                    `json:"project"`
	ProjectDir     string                    `json:"project_dir"`
	Date           string                    `json:"date"`
	Start          string                    `json:"start"`
	End            string                    `json:"end"`
	DurationMin    float64                   `json:"duration_min"`
	Cost           float64                   `json:"cost"`
	Messages       int                       `json:"messages"`
	UserMessages   int                       `json:"user_messages"`
	AssistMessages int                       `json:"assistant_messages"`
	InputTokens    int                       `json:"input_tokens"`
	OutputTokens   int                       `json:"output_tokens"`
	CacheRead      int                       `json:"cache_read_tokens"`
	CacheWrite     int                       `json:"cache_write_tokens"`
	APICalls       int                       `json:"api_calls"`
	PrimaryModel   string                    `json:"primary_model"`
	ModelBreakdown map[string]ModelBreakdown `json:"model_breakdown"`
	Tools          map[string]int            `json:"tools"`
	FirstPrompt    string                    `json:"first_prompt"`
	Slug           string                    `json:"slug"`
	FileSizeMB     float64                   `json:"file_size_mb"`
}

type ModelBreakdown struct {
	Cost         float64 `json:"cost"`
	OutputTokens int     `json:"output_tokens"`
	Calls        int     `json:"calls"`
}

type ToolCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Insights struct {
	Plans       []PlanFile      `json:"plans"`
	Plugins     PluginsData     `json:"plugins"`
	Todos       TodoStats       `json:"todos"`
	FileHistory FileHistoryData `json:"file_history"`
	Storage     StorageData     `json:"storage"`
}

type PlanFile struct {
	Filename string  `json:"filename"`
	Slug     string  `json:"slug"`
	Title    string  `json:"title"`
	Created  string  `json:"created"`
	Modified string  `json:"modified"`
	SizeKB   float64 `json:"size_kb"`
	Lines    int     `json:"lines"`
}

type PluginsData struct {
	Installed        []PluginInfo       `json:"installed"`
	Settings         PluginSettings     `json:"settings"`
	MarketplaceStats map[string]int     `json:"marketplace_stats"`
}

type PluginInfo struct {
	Name        string `json:"name"`
	ShortName   string `json:"short_name"`
	Marketplace string `json:"marketplace"`
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	LastUpdated string `json:"last_updated"`
}

type PluginSettings struct {
	PermissionMode string          `json:"permission_mode"`
	AutoUpdates    string          `json:"auto_updates"`
	EnabledPlugins map[string]bool `json:"enabled_plugins"`
}

type TodoStats struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Pending   int `json:"pending"`
	Files     int `json:"files"`
}

type FileHistoryData struct {
	TotalFiles    int     `json:"total_files"`
	TotalSessions int     `json:"total_sessions"`
	TotalSizeMB   float64 `json:"total_size_mb"`
}

type StorageData struct {
	TotalMB float64       `json:"total_mb"`
	Items   []StorageItem `json:"items"`
}

type StorageItem struct {
	Name   string  `json:"name"`
	SizeMB float64 `json:"size_mb"`
}

// Internal types for session parsing
type rawSession struct {
	SessionID     string
	ProjectDir    string
	ProjectPath   string
	Timestamps    []int64
	Models        map[string]*modelAccum
	Tools         map[string]int
	MessageCount  int
	UserMsgCount  int
	AssistMsgCount int
	FirstPrompt   string
	FileSize      int64
	Slug          string
	Source        string
}

type modelAccum struct {
	InputTokens    int
	OutputTokens   int
	CacheRead      int
	CacheCreation  int
	Cost           float64
	Calls          int
}

// Config types
type Config struct {
	Language    string        `json:"language"`
	PlanHistory []PlanConfig `json:"plan_history"`
	Migration   MigrationCfg `json:"migration"`
}

type PlanConfig struct {
	Plan       string   `json:"plan"`
	Start      string   `json:"start"`
	End        *string  `json:"end"`
	CostEUR    *float64 `json:"cost_eur"`
	CostUSD    float64  `json:"cost_usd"`
	BillingDay int      `json:"billing_day"`
}

type MigrationCfg struct {
	Enabled          bool    `json:"enabled"`
	Dir              *string `json:"dir"`
	ClaudeDirName    string  `json:"claude_dir_name"`
	DotClaudeJSONName string `json:"dot_claude_json_name"`
}
