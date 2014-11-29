package main

import (
	"fmt"
	"os"

	"github.com/fsouza/go-dockerclient"
)

type Tool struct {
	containers []Container
	ids        []string
}

type Container struct {
	Image        string
	PortBindings []string // e.g. 8000:8000. If empty exposes all
}

func WithContainers(images ...Container) *Tool {
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	imgs, err := client.ListImages(true)
	if err != nil {
		panic(err)
	}
	imgMap := make(map[string]bool)
	for _, img := range imgs {
		for _, tag := range img.RepoTags {
			imgMap[tag] = true
		}
	}

	for _, target := range images {
		if !imgMap[target.Image] {
			fmt.Printf("Pulling %s\n", target)
			req := docker.PullImageOptions{
				Repository:   target.Image,
				OutputStream: os.Stdout,
			}
			err := client.PullImage(req, docker.AuthConfiguration{})
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("%s image already present\n", target)
		}
	}

	for _, cont := range images {
		req := docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: cont.Image,
			},
		}
		c, err := client.CreateContainer(req)
		if err != nil {
			panic(err)
		}

		startConf := &docker.HostConfig{
			PublishAllPorts: true,
		}
		err = client.StartContainer(c.ID, startConf)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Started %s:%s", cont.Image, c.ID)
	}
	return nil
}

func main() {
	WithContainers(Container{Image: "dockerfile/rethinkdb:latest"}, Container{Image: "aglover/dynamodb-pier:latest"})
}
