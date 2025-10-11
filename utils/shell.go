package utils

import "os/exec"

//go:generate mockgen -source shell.go -destination shell_mock.go -package utils
type ShellExecutor interface {
	Command(name string, arg ...string) ([]byte, error)
}

type shell struct{}

func (s shell) Command(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	return cmd.CombinedOutput()
}

func NewShell() shell {
	return shell{}
}
