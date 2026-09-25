package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProjectFileInput 项目文件（多文件模式）
type ProjectFileInput struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// RunProject 执行多文件项目：写入临时目录后运行主文件。
// Go：main 包目录内 go run .（自动识别 main.go）；Python：运行 main.py。
func RunProject(language string, mainFile string, files []ProjectFileInput) *RunResult {
	if len(files) == 0 {
		return &RunResult{Error: "项目没有文件", ExitCode: -1}
	}
	for _, f := range files {
		if msg := rejectMalicious(f.Content); msg != "" {
			return &RunResult{Error: fmt.Sprintf("文件 %s: %s", f.Name, msg), ExitCode: -1}
		}
	}
	if err := sandboxBreaker.check(language); err != nil {
		return &RunResult{Error: err.Error(), ExitCode: -1}
	}

	dir, err := os.MkdirTemp("", "codelearn-project-*")
	if err != nil {
		return &RunResult{Error: "创建临时目录失败: " + err.Error(), ExitCode: -1}
	}
	defer os.RemoveAll(dir)

	for _, f := range files {
		name := strings.TrimSpace(f.Name)
		if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") {
			return &RunResult{Error: "非法文件名: " + name, ExitCode: -1}
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(f.Content), 0644); err != nil {
			return &RunResult{Error: "写入文件失败: " + err.Error(), ExitCode: -1}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	mainFile = strings.TrimSpace(mainFile)
	var cmd *exec.Cmd
	switch language {
	case "python", "py":
		if mainFile == "" {
			mainFile = "main.py"
		}
		if _, err := os.Stat(filepath.Join(dir, mainFile)); err != nil {
			return &RunResult{Error: "主文件不存在: " + mainFile, ExitCode: -1}
		}
		cmd = limitWrap(ctx, language, []string{"python", filepath.Join(dir, mainFile)})
	case "go":
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
			if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module codelearnproject\n\ngo 1.25\n"), 0644); err != nil {
				return &RunResult{Error: "写入 go.mod 失败: " + err.Error(), ExitCode: -1}
			}
		}
		cmd = limitWrap(ctx, language, []string{"go", "run", "."})
	default:
		return &RunResult{Error: "不支持的语言: " + language, ExitCode: -1}
	}
	cmd.Dir = dir

	r := execCmd(ctx, cmd, "")
	if strings.Contains(r.Error, "执行超时") {
		sandboxBreaker.recordFailure(language)
	} else if r.ExitCode == 0 {
		sandboxBreaker.recordSuccess(language)
	}
	return r
}
