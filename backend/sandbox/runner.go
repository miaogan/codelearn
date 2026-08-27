package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RunResult struct {
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exit_code"`
	TimeMs   int64  `json:"time_ms"`
}

type TestCaseResult struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Pass     bool   `json:"pass"`
	Error    string `json:"error,omitempty"`
}

type JudgeResult struct {
	Results    []TestCaseResult `json:"results"`
	AllPass    bool             `json:"all_pass"`
	PassCount  int              `json:"pass_count"`
	TotalCount int              `json:"total_count"`
}

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

const defaultTimeout = 10 * time.Second

// RunCode 执行用户提交的代码并返回输出。
// 注意：当前为直接执行，生产环境应使用 Docker 容器隔离。
func RunCode(language, code string) *RunResult {
	dir, err := os.MkdirTemp("", "codelearn-*")
	if err != nil {
		return &RunResult{Error: "创建临时目录失败: " + err.Error(), ExitCode: -1}
	}
	defer os.RemoveAll(dir)

	ext := fileExtension(language)
	if ext == "" {
		return &RunResult{Error: "不支持的语言: " + language, ExitCode: -1}
	}

	path := filepath.Join(dir, "main"+ext)
	if err := os.WriteFile(path, []byte(code), 0644); err != nil {
		return &RunResult{Error: "写入代码文件失败: " + err.Error(), ExitCode: -1}
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := buildRunCommand(ctx, language, path)
	if cmd == nil {
		return &RunResult{Error: "不支持的语言: " + language, ExitCode: -1}
	}

	return execCmd(ctx, cmd, "")
}

// JudgeCode 用测试用例评判用户代码
func JudgeCode(language, code string, testCases []TestCase) *JudgeResult {
	result := &JudgeResult{
		Results:    make([]TestCaseResult, 0, len(testCases)),
		TotalCount: len(testCases),
	}

	dir, err := os.MkdirTemp("", "codelearn-judge-*")
	if err != nil {
		for _, tc := range testCases {
			result.Results = append(result.Results, TestCaseResult{
				Input: tc.Input, Expected: tc.Expected, Error: "创建临时目录失败",
			})
		}
		return result
	}
	defer os.RemoveAll(dir)

	fullCode := wrapCode(language, code)
	ext := fileExtension(language)
	path := filepath.Join(dir, "main"+ext)
	if err := os.WriteFile(path, []byte(fullCode), 0644); err != nil {
		for _, tc := range testCases {
			result.Results = append(result.Results, TestCaseResult{
				Input: tc.Input, Expected: tc.Expected, Error: "写入代码文件失败",
			})
		}
		return result
	}

	// Python：直接用源文件逐个运行；Go：先编译一次再逐个运行
	binaryPath := ""
	if language == "go" {
		binaryPath = filepath.Join(dir, "main.exe")
		buildCtx, buildCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer buildCancel()
		buildCmd := exec.CommandContext(buildCtx, "go", "build", "-o", binaryPath, path)
		buildCmd.Dir = dir
		buildOut, buildErr := buildCmd.CombinedOutput()
		if buildErr != nil {
			for _, tc := range testCases {
				result.Results = append(result.Results, TestCaseResult{
					Input: tc.Input, Expected: tc.Expected,
					Error: "编译失败: " + string(buildOut),
				})
			}
			return result
		}
	}

	for _, tc := range testCases {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		var cmd *exec.Cmd
		if language == "go" && binaryPath != "" {
			cmd = exec.CommandContext(ctx, binaryPath)
		} else {
			cmd = buildRunCommand(ctx, language, path)
		}
		if cmd == nil {
			result.Results = append(result.Results, TestCaseResult{
				Input: tc.Input, Expected: tc.Expected, Error: "不支持的语言",
			})
			cancel()
			continue
		}

		r := execCmd(ctx, cmd, tc.Input)
		cancel()

		actual := strings.TrimSpace(r.Output)
		expected := strings.TrimSpace(tc.Expected)
		pass := actual == expected && r.ExitCode == 0

		tcResult := TestCaseResult{
			Input:    tc.Input,
			Expected: tc.Expected,
			Actual:   r.Output,
			Pass:     pass,
		}
		if r.Error != "" {
			tcResult.Error = r.Error
		}
		result.Results = append(result.Results, tcResult)
		if pass {
			result.PassCount++
		}
	}
	result.AllPass = result.PassCount == result.TotalCount
	return result
}

// JudgeCodeJSON 从 JSON 字符串解析测试用例后评判
func JudgeCodeJSON(language, code, testCasesJSON string) *JudgeResult {
	var testCases []TestCase
	if err := json.Unmarshal([]byte(testCasesJSON), &testCases); err != nil {
		return &JudgeResult{
			Results:    []TestCaseResult{{Error: "测试用例解析失败: " + err.Error()}},
			TotalCount: 0,
		}
	}
	if len(testCases) == 0 {
		return &JudgeResult{AllPass: false, TotalCount: 0}
	}
	return JudgeCode(language, code, testCases)
}

func execCmd(ctx context.Context, cmd *exec.Cmd, stdin string) *RunResult {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start).Milliseconds()

	result := &RunResult{
		Output:   stdout.String(),
		ExitCode: 0,
		TimeMs:   elapsed,
	}
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "执行超时（超过 " + fmt.Sprintf("%v", defaultTimeout) + "）"
			result.ExitCode = -1
		} else {
			errMsg := stderr.String()
			if errMsg == "" {
				errMsg = err.Error()
			}
			result.Error = errMsg
			result.ExitCode = 1
		}
	}
	return result
}

func buildRunCommand(ctx context.Context, language, path string) *exec.Cmd {
	switch language {
	case "python", "py":
		return exec.CommandContext(ctx, "python", path)
	case "go":
		return exec.CommandContext(ctx, "go", "run", path)
	default:
		return nil
	}
}

func fileExtension(language string) string {
	switch language {
	case "python", "py":
		return ".py"
	case "go":
		return ".go"
	default:
		return ""
	}
}

// wrapCode 为 Go 代码补充 package main 声明（如果用户省略了的话）
func wrapCode(language, code string) string {
	if language != "go" {
		return code
	}
	trimmed := strings.TrimSpace(code)
	if strings.HasPrefix(trimmed, "package ") {
		return code
	}
	return "package main\n\n" + code
}
