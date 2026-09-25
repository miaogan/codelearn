package sandbox

import (
	"strings"
	"testing"
	"time"
)

func TestRejectMalicious_Go(t *testing.T) {
	cases := []string{
		`import "os/exec"`,
		`import "syscall"`,
		`import "unsafe"`,
		`exec.Command("sh", "-c", "rm -rf /")`,
		`os.RemoveAll("/etc")`,
	}
	for _, code := range cases {
		if msg := rejectMalicious(code); msg == "" {
			t.Errorf("expected rejection for: %s", code)
		}
	}
}

func TestRejectMalicious_Python(t *testing.T) {
	cases := []string{
		`import subprocess`,
		`os.system("ls")`,
		`eval("__import__('os')")`,
		`__import__('socket')`,
		`shutil.rmtree("/")`,
	}
	for _, code := range cases {
		if msg := rejectMalicious(code); msg == "" {
			t.Errorf("expected rejection for: %s", code)
		}
	}
}

func TestRejectMalicious_SafeCode(t *testing.T) {
	safe := []string{
		`print("Hello")`,
		`package main
import "fmt"
func main() { fmt.Println("hi") }`,
		`n = int(input())
print(n * 2)`,
	}
	for _, code := range safe {
		if msg := rejectMalicious(code); msg != "" {
			t.Errorf("safe code rejected: %s (%s)", code, msg)
		}
	}
}

func TestRejectMalicious_TooLong(t *testing.T) {
	code := strings.Repeat("x", maxCodeLength+1)
	if msg := rejectMalicious(code); msg == "" {
		t.Error("expected rejection for oversized code")
	}
}

func TestRejectMaliciousProject_AllowsFileIO(t *testing.T) {
	allowed := []string{
		`os.WriteFile("out.txt", []byte("x"), 0644)`,
		`os.Create("data.txt")`,
		`os.OpenFile("log.txt", os.O_APPEND, 0644)`,
		`os.Remove("tmp.txt")`,
		`os.RemoveAll("build")`,
		`os.MkdirAll("dist", 0755)`,
		`open("data.txt", "w")`,
		`os.unlink("tmp.txt")`,
		`os.rmdir("dir")`,
	}
	for _, code := range allowed {
		if msg := rejectMaliciousProject(code); msg != "" {
			t.Errorf("project file I/O rejected: %s (%s)", code, msg)
		}
	}
}

func TestRejectMaliciousProject_StillBlocksDangerous(t *testing.T) {
	blocked := []string{
		`import "os/exec"`,
		`import "net/http"`,
		`import "syscall"`,
		`import "unsafe"`,
		`exec.Command("ls")`,
		`import subprocess`,
		`os.system("ls")`,
		`socket.socket()`,
		`import requests`,
		`__import__('os')`,
		`eval("2+2")`,
	}
	for _, code := range blocked {
		if msg := rejectMaliciousProject(code); msg == "" {
			t.Errorf("expected rejection in project mode: %s", code)
		}
	}
}

func TestCircuitBreaker_OpensAndBlocks(t *testing.T) {
	b := newCircuitBreaker()
	b.openFor = 10 * time.Second

	// 未失败时放行
	if err := b.check("python"); err != nil {
		t.Fatalf("expected allowed before failures, got %v", err)
	}

	// 连续失败达到阈值 → 熔断打开
	for i := 0; i < breakerMaxFailures; i++ {
		b.recordFailure("python")
	}
	if err := b.check("python"); err == nil {
		t.Error("expected breaker to be open after max failures")
	}

	// 其他语言不受影响
	if err := b.check("go"); err != nil {
		t.Errorf("expected go language unaffected, got %v", err)
	}
}

func TestCircuitBreaker_Recovers(t *testing.T) {
	b := newCircuitBreaker()
	b.openFor = 50 * time.Millisecond

	for i := 0; i < breakerMaxFailures; i++ {
		b.recordFailure("go")
	}
	if err := b.check("go"); err == nil {
		t.Fatal("expected breaker open")
	}

	time.Sleep(60 * time.Millisecond)
	if err := b.check("go"); err != nil {
		t.Errorf("expected breaker recovered, got %v", err)
	}
}

func TestCircuitBreaker_SuccessResets(t *testing.T) {
	b := newCircuitBreaker()
	b.openFor = 10 * time.Second

	b.recordFailure("python")
	b.recordFailure("python")
	b.recordSuccess("python")

	// 成功复位后继续计数：再失败 3 次即达阈值（4-1+3=6? 不，失败从 0 开始）
	for i := 0; i < breakerMaxFailures; i++ {
		b.recordFailure("python")
	}
	if err := b.check("python"); err == nil {
		t.Error("expected breaker open after reset failures")
	}
}

func TestShellQuote(t *testing.T) {
	if got := shellQuote("plain"); got != "'plain'" {
		t.Errorf("unexpected quote: %s", got)
	}
	if got := shellQuote("it's"); got != `'it'\''s'` {
		t.Errorf("unexpected quote with apostrophe: %s", got)
	}
}
