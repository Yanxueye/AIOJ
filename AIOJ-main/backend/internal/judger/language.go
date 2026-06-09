package judger

import "strings"

var supportedLanguages = map[string]bool{
	"cpp":    true,
	"python": true,
	"go":     true,
}

// NormalizeLanguage checks whether the AIOJ-facing language is supported by
// the current remote_judge integration.
func NormalizeLanguage(language string) (string, bool) {
	lang := strings.TrimSpace(strings.ToLower(language))
	_, ok := supportedLanguages[lang]
	return lang, ok
}
