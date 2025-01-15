package shell

type BashInput struct {
	Args       map[string]string
	StdinData  []byte
	WithStdout bool
	WithStderr bool
}

type BashOutput struct {
	Command    string
	ExitCode   int
	StdoutData []byte
	StderrData []byte
}
