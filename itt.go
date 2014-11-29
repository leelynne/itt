package itt

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

var client *docker.Client
var pulledContainers map[string]bool

func init() {
	endpoint := "unix:///var/run/docker.sock"
	nclient, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	client = nclient
	pulledContainers, err = getLocalContainers()
	if err != nil {
		panic(fmt.Sprintf("Failed to pull containers - %s", err.Error()))
	}
}

type Container struct {
	Image        string
	PortBindings []string // e.g. 8000:8000. If empty exposes all
}

func WithContainers(t *testing.T, names ...string) *Manager {
	m := Manager{}

	wg := sync.WaitGroup{}
	pulled := false
	for _, target := range names {
		if !pulledContainers[target] {
			wg.Add(1)
			pulled = true
			go func(cName string) {
				defer wg.Done()
				t.Logf("Pulling %s\n", cName)
				req := docker.PullImageOptions{
					Repository:   cName,
					OutputStream: os.Stdout,
				}
				err := client.PullImage(req, docker.AuthConfiguration{})
				if err != nil {
					t.Fatalf("Failed to pull image %s - %s", cName, err.Error())
				}
			}(target)
		}
	}
	wg.Wait()
	if pulled {
		newContainers, err := getLocalContainers()
		if err != nil {
			t.Fatalf("Failed to get local container list - %s", err.Error())
		}
		pulledContainers = newContainers
	}

	for _, target := range names {
		req := docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: target,
			},
		}
		c, err := client.CreateContainer(req)
		if err != nil {
			t.Fatalf("Failed to start container %s", target)
		}

		startConf := &docker.HostConfig{
			PublishAllPorts: true,
		}
		err = client.StartContainer(c.ID, startConf)
		if err != nil {
			panic(err)
		}
		m.ids = append(m.ids, c.ID)
		t.Logf("Started container %s:%s\n", target, c.ID)
	}

	return &m
}

func getLocalContainers() (map[string]bool, error) {
	imgs, err := client.ListImages(true)
	if err != nil {
		return nil, err
	}
	imgMap := make(map[string]bool)
	for _, img := range imgs {
		for _, tag := range img.RepoTags {
			imgMap[tag] = true
		}
	}
	return imgMap, nil
}
