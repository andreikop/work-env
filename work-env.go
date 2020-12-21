package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	// for docker attach "github.com/moby/moby/pkg/stdcopy"

	"github.com/alecthomas/kong"
)

func buildEnvinronment(client *client.Client, path, image string) error {
	command := exec.Command("/usr/bin/docker", "build", path, "--tag", image, "--label", "app=work-env")

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return fmt.Errorf("Failed to execute 'docker exec': %v", err)
	}

	return nil
}

func createWorkEnv(client *client.Client, image, name string) (containerId string, err error) {
	workDir, _ := os.Getwd()

	user, err := user.Current()
	userName := ""
	if err == nil {
		userName = user.Username
	} else {
		fmt.Printf("Failed to get current user: %v", err)
	}

	var containerConf = container.Config{
		Hostname:     name,
		User:         userName,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
		Env:          nil, // TODO
		Cmd:          nil, // TODO             strslice.StrSlice   // Command to run when starting the container
		Image:        image,
		Volumes:      nil,
		WorkingDir:   workDir,
		Entrypoint:   nil, //      strslice.StrSlice   // Entrypoint to run when starting the container
		Labels:       map[string]string{"app": "work-env"},
	}

	var hostConf = container.HostConfig{
		Mounts: []mount.Mount{
			{Type: "bind", Source: "/home", Target: "/home"},
			{Type: "bind", Source: "/dev", Target: "/dev"},
			{Type: "bind", Source: "/sys", Target: "/sys"},
			{Type: "bind", Source: "/tmp", Target: "/tmp"},
			{Type: "bind", Source: "/etc/resolv.conf", Target: "/etc/resolf.conf"},
		},
	}

	resp, err := client.ContainerCreate(
		context.Background(),
		&containerConf,
		&hostConf,
		nil, // networkingConfig
		name)
	if err != nil {
		return "", fmt.Errorf("Failed to start container: %v", err)
	}

	err = client.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", fmt.Errorf("Failed to start container: %v", err)
	}

	return resp.ID, nil
}

func attachToContainer(client *client.Client, containerName string) error {
	// TODO convert name to id
	/*
		hijackedCon, err := client.ContainerAttach(
			context.Background(),
			resp.ID,
			types.ContainerAttachOptions{
				Stdin: true,
				Stdout: true,
				Stderr: true,
			})
		if err != nil {
			return fmt.Errorf("Failed to attach to a container: %v", err)
		}

		defer hijackedCon.Close()

		written, err := stdcopy.StdCopy(os.Stdout, os.Stderr, hijackedCon.Reader)
		if err != nil {
			return fmt.Errorf("Failed to copy IO streams: %v", err)
		}
		fmt.Printf("~~~~ read done %d", written)
	*/
	// TODO use docker API instead of docker command

	// TODO do not hardcode the shell
	command := exec.Command("/usr/bin/docker", "exec", "-it", containerName, "zsh")

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Start()
	if err != nil {
		return fmt.Errorf("Failed to start 'docker exec': %v", err)
	}
	defer command.Wait()

	return nil
}

func removeContainer(client *client.Client, containerName string) error {
	return client.ContainerRemove(
		context.Background(),
		containerName,
		types.ContainerRemoveOptions{Force: true})
}

func listImages(client *client.Client) error {
	var filter filters.Args = filters.NewArgs()
	filter.Add("label", "app=work-env")

	imgSummaries, err := client.ImageList(
		context.Background(),
		types.ImageListOptions{Filters: filter})
	if err != nil {
		return fmt.Errorf("Failed to list envinronments: %v", err)
	}

	for _, imgSummary := range imgSummaries {
		fmt.Printf("%s\n", imgSummary.ID)
	}

	return nil
}

func main() {
	var CLI struct {
		Build struct {
			Path  string `arg help:"Docker image build path"`
			Image string `arg help:"Name of the envinronment image"`
			// DockerFile string `arg help:"DockerFile used to create envinronment" default:"DockerFile"`
		} `cmd help:"Build new envinronment image <image-name> from a DockerFile in a current directory"`

		Create struct {
			Image string `arg help:"Name of a Docker image used to create envinronment"`
			Name  string `arg name:"env-name" help:"Name of the new envinronment"`
			Rm    bool   `help:"Remove envinronment after session finished"`
		} `cmd help:"Create new envinronment instance <env-name> from docker image <image> and attach to it. Overwrites existing containers."`

		Attach struct {
			Name string `arg name:"env-name" help:"Envinronment name (docker container) to attach"`
			Rm   bool   `help:"Remove envinronment after session finished"`
		} `cmd help:"Start working in envinronment. Start a container and attach to it."`

		Remove struct {
			Name string `arg name:"env-name" help:"Envinronment to remove"`
		} `cmd help:"Remove an envinronment instance"`
		Images struct {
		} `cmd help:"List envinronment images"`
	}

	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	ctx := kong.Parse(&CLI)

	switch ctx.Command() {
	case "build <path> <image>":
		err := buildEnvinronment(client, CLI.Build.Path, CLI.Build.Image)
		if err != nil {
			fmt.Printf("Failed to build an envinronment image: %v\n", err)
		}
	case "create <image> <env-name>":
		_, err := createWorkEnv(client, CLI.Create.Image, CLI.Create.Name)
		if err != nil {
			fmt.Printf("Failed to create work envinronment: %v\n", err)
			return
		}
		err = attachToContainer(client, CLI.Create.Name)
		if err != nil {
			fmt.Printf("Failed to enter to envinronment: %v\n", err)
			return
		}
		if CLI.Create.Rm {
			err = removeContainer(client, CLI.Create.Name)
			if err != nil {
				fmt.Printf("Failed to remove container: %v\n", err)
			}
		}
	case "enter <env-name>":
		err := attachToContainer(client, CLI.Attach.Name)
		if err != nil {
			fmt.Printf("Failed to attach to envinronment: %v\n", err)
		}
	case "remove <env-name>":
		err := removeContainer(client, CLI.Remove.Name)
		if err != nil {
			fmt.Printf("Failed to remove container: %v\n", err)
		}
	case "images":
		err := listImages(client)
		if err != nil {
			fmt.Printf("Failed to list images: %v\n", err)
		}
	default:
		panic(ctx.Command())
	}

}
