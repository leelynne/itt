package itt

import "fmt"

type Manager struct {
	containers []Container
	ids        []string
}

func (m *Manager) Close() {
	go func() {
		for _, id := range m.ids {
			err := client.StopContainer(id, 100)
			fmt.Println(err)
		}
	}()
}
