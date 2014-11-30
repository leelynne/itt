package itt

import (
	"fmt"
	"os"
	"strings"
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
	Tag          string
	PortBindings []string // e.g. 8000:8000. If empty exposes all
	Init         func() error
}

func (c *Container) id() string {
	return fmt.Sprintf("%s:%s", c.Image, c.Tag)
}

func WithContainers(t *testing.T, names ...string) *Manager {
	fmt.Printf("Num names %d", len(names))
	var containers []Container
	for _, name := range names {

		c := Container{}
		parts := strings.Split(name, ":")
		c.Image = parts[0]
		if len(parts) > 1 {
			c.Tag = parts[1]
		} else {
			c.Tag = "latest"
		}
		containers = append(containers, c)
	}
	return WithContainerCfgs(t, containers...)
}

func WithContainerCfgs(t *testing.T, containers ...Container) *Manager {
	fmt.Printf("Num containers %d", len(containers))
	m := Manager{}

	wg := sync.WaitGroup{}
	pulled := false
	for _, loopC := range containers {
		fmt.Printf("Id %s\n", loopC.id())

		if have := pulledContainers[loopC.id()]; !have {
			wg.Add(1)
			pulled = true
			go func(c Container) {
				defer wg.Done()
				t.Logf("Pulling %s\n", c.id())
				req := docker.PullImageOptions{
					Repository:   c.Image,
					OutputStream: os.Stdout,
					Tag:          c.Tag,
				}
				err := client.PullImage(req, docker.AuthConfiguration{})
				if err != nil {
					t.Fatalf("Failed to pull image %s - %s", c.id(), err.Error())
				}
			}(loopC)
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

	for _, c := range containers {
		imgDetail, err := client.InspectImage(c.id())
		if err != nil {
			panic(err)
		}
		fmt.Printf("Ports! %+v\n", imgDetail.Config.ExposedPorts)
		fmt.Printf("Container Ports! %+v\n", imgDetail.ContainerConfig.ExposedPorts)
		req := docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: c.id(),
			},
		}
		dockerC, err := client.CreateContainer(req)
		if err != nil {
			t.Fatalf("Failed to start container %s", c.id())
		}

		startConf := &docker.HostConfig{
			PortBindings: make(map[docker.Port][]docker.PortBinding),
		}
		for port := range imgDetail.Config.ExposedPorts {
			startConf.PortBindings[port] = []docker.PortBinding{docker.PortBinding{HostPort: port.Port()}}
		}

		err = client.StartContainer(dockerC.ID, startConf)
		if err != nil {
			panic(err)
		}
		m.ids = append(m.ids, dockerC.ID)
		t.Logf("Started container %s:%s\n", c.id(), dockerC.ID)
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
			//			fmt.Printf("Tag %s\n:", tag)
			imgMap[tag] = true
		}
	}
	return imgMap, nil
}
