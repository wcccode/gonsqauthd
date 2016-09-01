package nsqauth

type Logger interface {
	Output(calldepth int, s string) error
}
