package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// ProjectFileInput 项目文件（多文件模式）
type ProjectFileInput struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// sandboxUID/sandboxGID 项目运行降权用户（nobody）
const (
	sandboxUID = 65534
	sandboxGID = 65534
)

// RunProject 执行多文件项目：写入临时目录后以 nobody 降权运行主文件。
// Go：main 包目录内 go run .（自动识别 main.go）；Python：运行 main.py。
func RunProject(language string, mainFile string, files []ProjectFileInput) *RunResult {
	if len(files) == 0 {
		return &RunResult{Error: "项目没有文件", ExitCode: -1}
	}
	for _, f := range files {
		if msg := rejectMaliciousProject(f.Content); msg != "" {
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

	// 以 root 运行时降权为 nobody：目录 0777 供 nobody 读写，子进程 setuid 为 nobody。
	// 非 root 开发环境不降权（进程本身已无特权），仍允许文件 I/O。
	dropPriv := os.Geteuid() == 0
	if dropPriv {
		if err := os.Chmod(dir, 0o777); err != nil {
			return &RunResult{Error: "设置临时目录权限失败: " + err.Error(), ExitCode: -1}
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

	if dropPriv {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{Uid: sandboxUID, Gid: sandboxGID},
		}
		// 最小化降权进程环境：仅保留 PATH 与 HOME（供 mise/pyenv shim 解析真实工具），
		// 不透传父进程环境，避免数据库凭据等敏感信息暴露给被运行代码。
		home, _ := os.UserHomeDir()
		if home == "" {
			home = "/root"
		}
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home}
		if language == "go" {
			// nobody 无法写 root 家目录下的缓存，全部隔离到项目临时目录
			// GOTMPDIR 必须预先存在（go 只 stat 不创建），三个目录均建为 0777 供 nobody 写入。
			// 注意 MkdirAll 受 umask 影响，需再显式 chmod 0777。
			for _, d := range []string{".gocache", ".gopath", ".gotmp"} {
				if err := os.MkdirAll(filepath.Join(dir, d), 0o777); err != nil {
					return &RunResult{Error: "创建缓存目录失败: " + err.Error(), ExitCode: -1}
				}
				if err := os.Chmod(filepath.Join(dir, d), 0o777); err != nil {
					return &RunResult{Error: "设置缓存目录权限失败: " + err.Error(), ExitCode: -1}
				}
			}
			cmd.Env = append(cmd.Env,
				"GOMEMLIMIT=256MiB",
				"GOCACHE="+filepath.Join(dir, ".gocache"),
				"GOPATH="+filepath.Join(dir, ".gopath"),
				"GOTMPDIR="+filepath.Join(dir, ".gotmp"),
				"GOENV=off",
				"GOTOOLCHAIN=local",
			)
		}
	}

	r := execCmd(ctx, cmd, "")
	if strings.Contains(r.Error, "执行超时") {
		sandboxBreaker.recordFailure(language)
	} else if r.ExitCode == 0 {
		sandboxBreaker.recordSuccess(language)
	}
	return r
}
