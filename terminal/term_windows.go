package terminal

import (
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/ActiveState/termtest/conpty"
	"github.com/Uzer007-admin/granio/v2"
	"golang.org/x/sys/windows"
)

type windowsPTY struct {
	*conpty.ConPty
	lock   sync.Mutex
	once   sync.Once
	closed bool
	err    error
}

func (p *windowsPTY) Close() error {
	p.once.Do(func() {
		_ = p.InPipe().Close()
		_ = p.OutPipe().Close()
		p.lock.Lock()
		defer p.lock.Unlock()
		p.closed = true
		p.err = p.ConPty.Close()
	})
	return p.err
}

func (p *windowsPTY) resize(columns, rows uint16) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.closed {
		_ = p.Resize(columns, rows)
	}
}

func (t *Terminal) updatePTYSize() {
	t.closeLock.Lock()
	defer t.closeLock.Unlock()
	if t.pty == nil { // during load
		return
	}
	t.pty.(*windowsPTY).resize(uint16(t.config.Columns), uint16(t.config.Rows))
}

// hangup does nothing on Windows, closing the console will end its processes.
func (t *Terminal) hangup() {
}

func (t *Terminal) startPTY() (io.WriteCloser, io.Reader, io.Closer, error) {
	cpty, err := conpty.New(80, 25)
	if err != nil {
		return nil, nil, nil, err
	}

	pty := &windowsPTY{ConPty: cpty}
	pid, processHandle, err := cpty.Spawn(
		"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		[]string{},
		&syscall.ProcAttr{
			Env: os.Environ(),
			Dir: t.startingDir(),
		},
	)
	if err != nil {
		_ = pty.Close()
		return nil, nil, nil, err
	}
	defer windows.CloseHandle(windows.Handle(processHandle))

	t.cmd = &exec.Cmd{}
	t.cmd.Dir = t.startingDir()
	process, err := os.FindProcess(pid)
	if err != nil {
		_ = pty.Close()
		return nil, nil, nil, err
	}
	go func() {
		ps, err := process.Wait()
		if err != nil {
			granio.LogError("Error waiting for terminal process", err)
		} else {
			t.exitStatus.Store(int64(ps.ExitCode()) + 1)
		}
		_ = pty.Close()
	}()

	return cpty.InPipe(), cpty.OutPipe(), pty, nil
}
