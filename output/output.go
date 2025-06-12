package output

type Writer interface {
	Name() string
	Send(content, sessionID, producerName, producerType string) error
	WithIgnored(producers ...string) Writer
}
