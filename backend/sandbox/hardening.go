package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ===== 恶意代码检测 =====

const maxCodeLength = 20000

// alwaysDangerous 单文件与项目模式都必须拦截的危险特征：
// 子进程/系统调用/网络，降权运行也无法阻止这些行为（nobody 仍可发起网络请求），需静态拦截。
var alwaysDangerous = []string{
	// Go
	"os/exec", "syscall", "unsafe", "exec.command", "net.", "net/",
	"crypto/", "plugin.", "go:linkname",
	// Python
	"subprocess", "os.system", "os.popen", "socket", "import requests",
	"requests.", "urllib", "http.server", "http.client", "importlib",
	"eval(", "exec(", "__import__", "pickle.loads", "getattr", "ctypes",
	"ftplib", "telnetlib", "smtplib", "marshal.loads",
	// 通用
	"rm -rf", "shutdown", "mkfs", "dd if=",
}

// fileIOPatterns 仅在单文件模式（root 直接运行）额外拦截的文件操作特征。
// 项目模式以 nobody 降权运行，这些操作无法越权写系统目录，故允许。
var fileIOPatterns = []string{
	"os.remove", "os.chmod", "os.writefile", "os.create", "os.openfile",
	"os.unlink", "os.rmdir", "shutil.rmtree",
}

// rejectMalicious 单文件代码严格检查（含文件操作拦截）。
func rejectMalicious(code string) string { return reject(code, false) }

// rejectMaliciousProject 项目模式检查：允许文件 I/O（降权运行保证安全），
// 但仍拦截网络/子进程/系统调用。
func rejectMaliciousProject(code string) string { return reject(code, true) }

func reject(code string, allowFileIO bool) string {
	if len(code) > maxCodeLength {
		return fmt.Sprintf("代码长度超过限制（%d 字符）", maxCodeLength)
	}
	lower := strings.ToLower(code)
	for _, p := range alwaysDangerous {
		if strings.Contains(lower, p) {
			return fmt.Sprintf("检测到潜在危险代码（%s），已拦截", p)
		}
	}
	if !allowFileIO {
		for _, p := range fileIOPatterns {
			if strings.Contains(lower, p) {
				return fmt.Sprintf("检测到潜在危险代码（%s），已拦截", p)
			}
		}
	}
	return ""
}

// ===== 异常熔断 =====

const (
	breakerMaxFailures = 5
	breakerOpenFor     = 30 * time.Second
)

// circuitBreaker 按语言维度的简单熔断器：连续失败达到阈值后短暂打开
type circuitBreaker struct {
	mu        sync.Mutex
	failures  map[string]int
	openUntil map[string]time.Time
	openFor   time.Duration
}

func newCircuitBreaker() *circuitBreaker {
	return &circuitBreaker{
		failures:  map[string]int{},
		openUntil: map[string]time.Time{},
		openFor:   breakerOpenFor,
	}
}

func (b *circuitBreaker) check(language string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if until, ok := b.openUntil[language]; ok && time.Now().Before(until) {
		return fmt.Errorf("判题服务繁忙，请稍后再试")
	}
	return nil
}

func (b *circuitBreaker) recordFailure(language string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures[language]++
	if b.failures[language] >= breakerMaxFailures {
		b.openUntil[language] = time.Now().Add(b.openFor)
		b.failures[language] = 0
	}
}

func (b *circuitBreaker) recordSuccess(language string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures[language] = 0
}

var sandboxBreaker = newCircuitBreaker()

// ===== 资源限额 =====

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// limitWrap 用 bash ulimit 包裹执行命令，限制 CPU 时间/文件大小/进程数。
// Go 运行时虚拟地址预留较大，无法用 ulimit -v 限制，改用 GOMEMLIMIT 软上限；
// Python 直接限制虚拟内存 256MB。
func limitWrap(ctx context.Context, language string, argv []string) *exec.Cmd {
	quoted := make([]string, len(argv))
	for i, a := range argv {
		quoted[i] = shellQuote(a)
	}
	inner := strings.Join(quoted, " ")

	// Go 冷编译标准库会写大文件（构建缓存最大约 45MB），需放宽单文件大小上限；Python 保持较小值
	fileLimit := uint64(8192) // 4MB（512 字节块）
	if language == "go" {
		fileLimit = 262144 // 128MB
	}
	limits := fmt.Sprintf("ulimit -t 8; ulimit -f %d; ulimit -u 64;", fileLimit)
	if language == "go" {
		cmd := exec.CommandContext(ctx, "bash", "-c", limits+" exec "+inner)
		cmd.Env = append(os.Environ(), "GOMEMLIMIT=256MiB")
		return cmd
	}
	cmd := exec.CommandContext(ctx, "bash", "-c", limits+" ulimit -v 262144; exec "+inner)
	cmd.Env = append(os.Environ(), "HOME=/nonexistent")
	return cmd
}
