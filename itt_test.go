package itt_test

import (
	"testing"

	"github.com/talio/itt"
)

func TestRethink(t *testing.T) {
	c := itt.WithContainers(t, "dockerfile/rethinkdb:latest")
	defer c.Close()
}

func TestRethink2(t *testing.T) {
	c := itt.WithContainers(t, "dockerfile/rethinkdb:latest")
	defer c.Close()
}

func TestRethink3(t *testing.T) {
	c := itt.WithContainers(t, "dockerfile/rethinkdb:latest")
	defer c.Close()
}

func TestRethink4(t *testing.T) {
	c := itt.WithContainers(t, "dockerfile/rethinkdb:latest")
	defer c.Close()

}
