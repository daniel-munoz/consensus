package output

type Writer interface {
	Send(content, sessionID, context string) error
	Name() string
}
