package sandbox

import (
	"strings"
	"testing"
)

func TestRunCode_Python(t *testing.T) {
	code := `print("Hello, World!")`
	result := RunCode("python", code)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	output := strings.TrimSpace(result.Output)
	if output != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got '%s'", result.Output)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestRunCode_PythonArithmetic(t *testing.T) {
	code := `print(5 * 5)`
	result := RunCode("python", code)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	output := strings.TrimSpace(result.Output)
	if output != "25" {
		t.Errorf("expected '25', got '%s'", result.Output)
	}
}

func TestRunCode_Go(t *testing.T) {
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello from Go")
}`
	result := RunCode("go", code)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	output := strings.TrimSpace(result.Output)
	if output != "Hello from Go" {
		t.Errorf("expected 'Hello from Go', got '%s'", result.Output)
	}
}

func TestRunCode_GoNoPackage(t *testing.T) {
	code := `func main() {
	print("test")
}`
	result := RunCode("go", code)

	if result.Error == "" {
		t.Error("expected error for Go code without package declaration")
	}
}

func TestRunCode_UnsupportedLanguage(t *testing.T) {
	result := RunCode("ruby", "puts 'hello'")

	if result.Error == "" {
		t.Error("expected error for unsupported language")
	}
	if result.Error != "不支持的语言: ruby" {
		t.Errorf("expected unsupported language error, got: %s", result.Error)
	}
}

func TestRunCode_PythonSyntaxError(t *testing.T) {
	code := `print("unterminated`
	result := RunCode("python", code)

	if result.Error == "" {
		t.Error("expected error for syntax error")
	}
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
}

func TestRunCode_EmptyCode(t *testing.T) {
	result := RunCode("python", "")

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0 for empty Python, got %d", result.ExitCode)
	}
}

func TestJudgeCode_PythonAllPass(t *testing.T) {
	code := `n = int(input())
print(n * 2)`
	testCases := []TestCase{
		{Input: "3", Expected: "6"},
		{Input: "5", Expected: "10"},
		{Input: "0", Expected: "0"},
	}

	result := JudgeCode("python", code, testCases)

	if !result.AllPass {
		t.Error("expected all tests to pass")
	}
	if result.PassCount != 3 {
		t.Errorf("expected 3 passes, got %d", result.PassCount)
	}
	if result.TotalCount != 3 {
		t.Errorf("expected 3 total, got %d", result.TotalCount)
	}
}

func TestJudgeCode_PythonPartialPass(t *testing.T) {
	code := `n = int(input())
print(n * 2)`
	testCases := []TestCase{
		{Input: "3", Expected: "6"},
		{Input: "5", Expected: "11"},
		{Input: "0", Expected: "0"},
	}

	result := JudgeCode("python", code, testCases)

	if result.AllPass {
		t.Error("expected not all pass")
	}
	if result.PassCount != 2 {
		t.Errorf("expected 2 passes, got %d", result.PassCount)
	}
}

func TestJudgeCode_GoAllPass(t *testing.T) {
	code := `package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)
	fmt.Println(n * 2)
}`
	testCases := []TestCase{
		{Input: "3", Expected: "6"},
		{Input: "7", Expected: "14"},
	}

	result := JudgeCode("go", code, testCases)

	if !result.AllPass {
		t.Errorf("expected all pass, got %d/%d", result.PassCount, result.TotalCount)
	}
}

func TestJudgeCode_GoNoPackage(t *testing.T) {
	code := `func main() {
	print("test")
}`
	testCases := []TestCase{
		{Input: "", Expected: "test"},
	}

	result := JudgeCode("go", code, testCases)

	if result.AllPass {
		t.Error("expected failure for Go code without package")
	}
	if result.PassCount != 0 {
		t.Errorf("expected 0 passes, got %d", result.PassCount)
	}
}

func TestJudgeCode_EmptyTestCases(t *testing.T) {
	result := JudgeCode("python", "print('hi')", []TestCase{})

	if result.TotalCount != 0 {
		t.Errorf("expected 0 total, got %d", result.TotalCount)
	}
	if len(result.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(result.Results))
	}
}

func TestJudgeCodeJSON_Valid(t *testing.T) {
	code := `n = int(input())
print(n + 1)`
	jsonStr := `[{"input":"5","expected":"6"},{"input":"10","expected":"11"}]`

	result := JudgeCodeJSON("python", code, jsonStr)

	if !result.AllPass {
		t.Errorf("expected all pass, got %d/%d", result.PassCount, result.TotalCount)
	}
}

func TestJudgeCodeJSON_InvalidJSON(t *testing.T) {
	result := JudgeCodeJSON("python", "print('hi')", "invalid json")

	if result.TotalCount != 0 {
		t.Errorf("expected 0 total for invalid JSON, got %d", result.TotalCount)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].Error == "" {
		t.Error("expected error message for invalid JSON")
	}
}

func TestFileExtension(t *testing.T) {
	tests := []struct {
		lang string
		ext  string
	}{
		{"python", ".py"},
		{"py", ".py"},
		{"go", ".go"},
		{"ruby", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := fileExtension(tt.lang)
		if got != tt.ext {
			t.Errorf("fileExtension(%q) = %q, want %q", tt.lang, got, tt.ext)
		}
	}
}

func TestWrapCode_Go(t *testing.T) {
	t.Run("adds package main when missing", func(t *testing.T) {
		code := `func main() {}`
		result := wrapCode("go", code)
		if result != "package main\n\nfunc main() {}" {
			t.Errorf("unexpected wrapped code: %s", result)
		}
	})

	t.Run("preserves code with package", func(t *testing.T) {
		code := `package main

func main() {}`
		result := wrapCode("go", code)
		if result != code {
			t.Errorf("should not modify code with package: %s", result)
		}
	})

	t.Run("does not wrap Python", func(t *testing.T) {
		code := `print("hello")`
		result := wrapCode("python", code)
		if result != code {
			t.Errorf("should not wrap Python code")
		}
	})
}
