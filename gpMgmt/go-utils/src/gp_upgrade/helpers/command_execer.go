package helpers

type CommandExecer func(name string, args ...string) Command

type Command interface {
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
	Start() error
	Run() error
}
