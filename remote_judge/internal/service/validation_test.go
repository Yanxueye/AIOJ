package service

import (
	"strings"
	"testing"

	"remote_judge/internal/domain"
)

// TestValidateCreateRequestNegativeUserID verifies UserID <= 0 rejection.
func TestValidateCreateRequestNegativeUserID(t *testing.T) {
	t.Logf(">>> Validation: UserID <= 0 -> ErrBadRequest")
	err := validateCreateRequest(CreateSubmissionRequest{
		UserID: 0, ProblemID: 1001, Language: "cpp17", Code: "int main(){}",
	})
	t.Logf("    UserID=0 -> error=%v (is ErrBadRequest=%v)", err, err == ErrBadRequest)
	if err == nil {
		t.Fatal("expected error for zero UserID")
	}
}

// TestValidateCreateRequestNegativeProblemID verifies ProblemID <= 0 rejection.
func TestValidateCreateRequestNegativeProblemID(t *testing.T) {
	t.Logf(">>> Validation: ProblemID <= 0 -> ErrBadRequest")
	err := validateCreateRequest(CreateSubmissionRequest{
		UserID: 1, ProblemID: 0, Language: "cpp17", Code: "int main(){}",
	})
	t.Logf("    ProblemID=0 -> error=%v", err)
	if err == nil {
		t.Fatal("expected error for zero ProblemID")
	}
}

// TestValidateCreateRequestBlankCode verifies empty or whitespace-only code rejection.
func TestValidateCreateRequestBlankCode(t *testing.T) {
	t.Logf(">>> Validation: blank code -> ErrBadRequest")
	for _, code := range []string{"", "   ", "\t\n "} {
		err := validateCreateRequest(CreateSubmissionRequest{
			UserID: 1, ProblemID: 1001, Language: "cpp17", Code: code,
		})
		t.Logf("    code=%q -> error=%v (is ErrBadRequest=%v)", code, err, err != nil && strings.Contains(err.Error(), "code required"))
		if err == nil {
			t.Fatalf("expected error for blank code %q", code)
		}
	}
}

// TestValidateCreateRequestCodeTooLong verifies code exceeding 128KB.
func TestValidateCreateRequestCodeTooLong(t *testing.T) {
	t.Logf(">>> Validation: code > 128KB -> ErrBadRequest")
	err := validateCreateRequest(CreateSubmissionRequest{
		UserID: 1, ProblemID: 1001, Language: "cpp17",
		Code: strings.Repeat("x", 128*1024+1),
	})
	t.Logf("    code_length=%d -> error=%v", 128*1024+1, err)
	if err == nil {
		t.Fatal("expected error for code too long")
	}
}

// TestValidateCreateRequestUnsupportedLanguage verifies language validation.
func TestValidateCreateRequestUnsupportedLanguage(t *testing.T) {
	t.Logf(">>> Validation: unsupported language -> ErrBadRequest")
	err := validateCreateRequest(CreateSubmissionRequest{
		UserID: 1, ProblemID: 1001, Language: "ruby", Code: "puts 1",
	})
	t.Logf("    language=ruby -> error=%v", err)
	if err == nil {
		t.Fatal("expected error for unsupported language")
	}
}

// TestValidateCreateRequestSupportedLanguages verifies all known languages pass.
func TestValidateCreateRequestSupportedLanguages(t *testing.T) {
	t.Logf(">>> Validation: all supported languages pass")
	for lang := range domain.SupportedLanguages {
		err := validateCreateRequest(CreateSubmissionRequest{
			UserID: 1, ProblemID: 1001, Language: lang, Code: "// ok",
		})
		if err != nil {
			t.Fatalf("unexpected error for language %s: %v", lang, err)
		}
	}
	t.Logf("    all %d languages validated ok", len(domain.SupportedLanguages))
}

// TestValidateCreateRequestExactMaxLength verifies code at exactly 128KB passes.
func TestValidateCreateRequestExactMaxLength(t *testing.T) {
	t.Logf(">>> Validation: code at exactly 128KB -> ok")
	err := validateCreateRequest(CreateSubmissionRequest{
		UserID: 1, ProblemID: 1001, Language: "cpp17",
		Code: strings.Repeat("a", 128*1024),
	})
	t.Logf("    code_length=%d -> err=%v", 128*1024, err)
	if err != nil {
		t.Fatalf("expected no error for max-length code, got %v", err)
	}
}
