package tools

// ModeConfig defines which tools a mode can use and the max rounds of tool calling.
type ModeConfig struct {
	ToolNames []string // names of tools from Definitions
	MaxRounds int      // max tool-calling rounds (0 = no tools)
}

// ModeTools maps mode strings to their configurations.
var ModeTools = map[string]ModeConfig{
	"chat": {
		ToolNames: []string{"search_problems", "query_user_problems", "retrieve_knowledge", "get_user_code"},
		MaxRounds: 3,
	},
	"code-diagnosis": {
		ToolNames: nil,
		MaxRounds: 0,
	},
	"generate-solution": {
		ToolNames: nil,
		MaxRounds: 0,
	},
	"knowledge-graph": {
		ToolNames: []string{"search_problems", "query_user_problems"},
		MaxRounds: 1,
	},
	"study-plan": {
		ToolNames: []string{"search_problems", "query_user_problems"},
		MaxRounds: 2,
	},
	"solve": {
		ToolNames: []string{"search_problems", "query_user_problems", "submit_code", "retrieve_knowledge", "get_user_code"},
		MaxRounds: 3,
	},
}

// ForMode returns the ToolDefs for a given mode.
func ForMode(mode string) []ToolDef {
	cfg, ok := ModeTools[mode]
	if !ok || len(cfg.ToolNames) == 0 {
		return nil
	}
	defs := make([]ToolDef, 0, len(cfg.ToolNames))
	for _, name := range cfg.ToolNames {
		if d, ok := Definitions[name]; ok {
			defs = append(defs, d)
		}
	}
	return defs
}

// MaxRoundsForMode returns the max tool-calling rounds for a mode.
func MaxRoundsForMode(mode string) int {
	cfg, ok := ModeTools[mode]
	if !ok {
		return 0
	}
	return cfg.MaxRounds
}

// IsValidMode returns true if the mode is known.
func IsValidMode(mode string) bool {
	_, ok := ModeTools[mode]
	return ok
}
