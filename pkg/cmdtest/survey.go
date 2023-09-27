package cmdtest

import (
	"github.com/Netflix/go-expect"
	"github.com/creack/pty"
	"github.com/hinshun/vt10x"
	"os"
	"testing"
)

type PromptTest struct {
	Name           string
	Procedure      func(*Console)
	CheckProcedure func() error
}

type Console struct {
	console *expect.Console
	t       *testing.T
}

func (c *Console) ExpectString(s string) {
	if _, err := c.console.ExpectString(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("ExpectString(%q) = %v", s, err)
	}
}

func (c *Console) SendLine(s string) {
	if _, err := c.console.SendLine(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("SendLine(%q) = %v", s, err)
	}
}

func (c *Console) Send(s string) {
	if _, err := c.console.Send(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("Send(%q) = %v", s, err)
	}
}

func (c *Console) ExpectEOF() {
	if _, err := c.console.ExpectEOF(); err != nil {
		c.t.Helper()
		c.t.Fatalf("ExpectEOF() = %v", err)
	}
}

func RunTest(t *testing.T, procedure func(*Console), test func() error) {
	t.Helper()

	pty, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("failed to open pseudotty: %v", err)
	}

	term := vt10x.New(vt10x.WithWriter(tty))
	c, err := expect.NewConsole(expect.WithStdin(pty), expect.WithStdout(term), expect.WithCloser(pty, tty))
	if err != nil {
		t.Fatalf("failed to create console: %v", err)
	}
	defer func(c *expect.Console) {
		err := c.Close()
		if err != nil {
			t.Error(err)
		}
	}(c)

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		procedure(&Console{console: c, t: t})
	}()

	// Backup the original stdin
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()
	originalStderr := os.Stderr
	defer func() { os.Stderr = originalStderr }()

	// Replace stdin, stdout, and stderr with test pipe
	os.Stdin = c.Tty()
	os.Stdout = c.Tty()
	os.Stderr = c.Tty()

	if err := test(); err != nil {
		t.Error(err)
	}

	if err := c.Tty().Close(); err != nil {
		t.Errorf("error closing Tty: %v", err)
	}

	// Restore stdin, stdout, and stderr
	os.Stdin = originalStdin
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	<-donec
}
