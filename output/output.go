package output

type Writer interface {
	Name() string
	Send(content, sessionID, producer string) error
	WithIgnored(producers ...string) Writer
}
