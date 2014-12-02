package itt

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

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
	Name         string
	PortBindings []string      // e.g. 8000:8000. If empty exposes all
	Init         func() error  // Function to run after the container is started and the Delay has passed
	Delay        time.Duration // Length of time to wait before the container is considered ready for use
	RandomPorts  bool          // True to map exposed ports to randomized ports

	image string
	tag   string
}

func (c *Container) id() string {
	return fmt.Sprintf("%s:%s", c.image, c.tag)
}

func WithContainers(t *testing.T, names ...string) *Manager {
	var containers []Container
	for _, name := range names {
		c := Container{
			Name:        name,
			Delay:       time.Duration(time.Millisecond * 200),
			RandomPorts: true,
		}
		parts := strings.Split(name, ":")
		c.image = parts[0]
		if len(parts) > 1 {
			c.tag = parts[1]
		} else {
			c.tag = "latest"
		}
		containers = append(containers, c)
	}
	return WithContainerCfgs(t, containers...)
}

func WithContainerCfgs(t *testing.T, containers ...Container) *Manager {
	m := Manager{
		t:            t,
		PortMappings: make(map[string]string),
	}

	wg := sync.WaitGroup{}
	pulled := false
	for _, loopC := range containers {
		if have := pulledContainers[loopC.id()]; !have {
			wg.Add(1)
			pulled = true
			go func(c Container) {
				defer wg.Done()
				t.Logf("Pulling %s\n", c.id())
				req := docker.PullImageOptions{
					Repository:   c.image,
					Tag:          c.tag,
					OutputStream: os.Stdout,
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
			hostPort := port.Port()
			if c.RandomPorts {
				hostPort = randomPort()
			}
			m.PortMappings[port.Port()] = hostPort
			startConf.PortBindings[port] = []docker.PortBinding{docker.PortBinding{HostPort: hostPort}}
		}

		err = client.StartContainer(dockerC.ID, startConf)
		if err != nil {
			panic(err)
		}
		m.ids = append(m.ids, dockerC.ID)
		t.Logf("Started container %s:%s\n", c.id(), dockerC.ID)

	}

	time.Sleep(time.Millisecond * 1000)
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

func randomPort() string {
	min := 1024
	max := 65535
	return strconv.Itoa(rand.Intn(max-min) + min)
}
