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

// dangerousPatterns 危险代码特征黑名单（大小写不敏感子串匹配）
var dangerousPatterns = []string{
	// Go
	"os/exec", "syscall", "unsafe", "exec.command", "os.remove", "os.chmod",
	"os.writefile", "os.create", "os.openfile", "net.", "crypto/", "plugin.",
	"reflect", "go:linkname",
	// Python
	"subprocess", "os.system", "os.popen", "os.remove", "os.unlink", "os.rmdir",
	"shutil.rmtree", "socket", "import requests", "urllib", "http.server",
	"importlib", "eval(", "exec(", "__import__", "pickle.loads", "getattr",
	"ctypes", "ftplib", "telnetlib", "smtplib", "marshal.loads",
	// 通用
	"rm -rf", "shutdown", "mkfs", "dd if=",
}

// rejectMalicious 静态检查代码是否含危险特征，返回拒绝原因；安全则返回空字符串
func rejectMalicious(code string) string {
	if len(code) > maxCodeLength {
		return fmt.Sprintf("代码长度超过限制（%d 字符）", maxCodeLength)
	}
	lower := strings.ToLower(code)
	for _, p := range dangerousPatterns {
		if strings.Contains(lower, p) {
			return fmt.Sprintf("检测到潜在危险代码（%s），已拦截", p)
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

	limits := "ulimit -t 8; ulimit -f 8192; ulimit -u 64;"
	cmd := exec.CommandContext(ctx, "bash", "-c", limits+" exec "+inner)
	if language == "go" {
		cmd.Env = append(os.Environ(), "GOMEMLIMIT=256MiB")
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", limits+" ulimit -v 262144; exec "+inner)
	}
	return cmd
}
