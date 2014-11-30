package itt

import "fmt"

type Manager struct {
	containers []Container
	ids        []string
}

func (m *Manager) Close() {

	fmt.Printf("Starting close on %d containers\n", len(m.ids))
	/*
		for _, id := range m.ids {
			err := client.KillContainer(docker.KillContainerOptions{ID: id})
			if err != nil {
				fmt.Println(err)
			}
		}*/
}
