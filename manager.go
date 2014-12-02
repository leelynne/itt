package itt

import (
	"bytes"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

type Manager struct {
	PortMappings map[string]string // Randomized port mappings from exposed port -> random
	containers   []Container
	ids          []string
	t            *testing.T
}

func (m *Manager) Fatal(args ...interface{}) {
	m.Close()
	m.getLogs()
	m.t.Fatal(args)
}

func (m *Manager) Fatalf(format string, args ...interface{}) {
	m.Close()
	m.t.Fatalf(format, args)
}

func (m *Manager) FailNow() {
	m.Close()
	m.getLogs()
	m.t.FailNow()
}
func (m *Manager) Close() {
	m.t.Logf("Starting close on %d containers\n", len(m.ids))
	for _, id := range m.ids {
		err := client.KillContainer(docker.KillContainerOptions{ID: id})
		if err != nil {
			m.t.Error(err)
		}
	}
}
func (m *Manager) Write(p []byte) (int, error) {
	m.t.Log("Writing!")
	m.t.Log(p)
	return len(p), nil
}
func (m *Manager) getLogs() {
	for _, id := range m.ids {
		buf := bytes.Buffer{}
		//		go func(cid string) {
		logReq := docker.LogsOptions{
			Container:    id,
			OutputStream: &buf,
			ErrorStream:  &buf,
			Stdout:       true,
			Stderr:       true,
			Timestamps:   true,
		}
		err := client.Logs(logReq)
		if err != nil {
			m.t.Error(err)
			return
		}
		m.t.Errorf("%s:\n%s", id, buf.String())
		//		}(id)
	}
}
