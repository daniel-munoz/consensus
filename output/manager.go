package output

import (
	"fmt"
	"os"
	"sync"
)

type Manager struct {
	outputs []Writer
}

func NewManager(outputs ...Writer) *Manager {
	return &Manager{outputs: outputs}
}

func (m *Manager) AddOutput(output Writer) {
	m.outputs = append(m.outputs, output)
}

func (m *Manager) Send(content, sessionID, context string) {
	var wg sync.WaitGroup

	for _, output := range m.outputs {
		wg.Add(1)
		go func(o Writer) {
			defer wg.Done()
			if err := o.Send(content, sessionID, context); err != nil {
				fmt.Fprintf(os.Stderr, "Error sending to %s output: %v\n", o.Name(), err)
			}
		}(output)
	}

	wg.Wait()
}

