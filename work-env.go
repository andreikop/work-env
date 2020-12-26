package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	// for docker attach "github.com/moby/moby/pkg/stdcopy"

	"github.com/alecthomas/kong"
)

func buildEnvironment(client *client.Client, path, image string) error {
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

func printImage(imgSummary *types.ImageSummary) {
	if len(imgSummary.RepoTags) > 0 {
		for _, repoTag := range imgSummary.RepoTags {
			if strings.Count(repoTag, ":") == 1 { // name:version
				parts := strings.Split(repoTag, ":")
				name := parts[0]
				version := parts[1]
				if version == "latest" {
					fmt.Printf("%s\n", name)
				} else {
					fmt.Printf("%s:%s\n", name, version)
				}
			} else {
				fmt.Printf("%s\n", repoTag)
			}
		}
	} else { // Strange, no tag. Let's show at least ID
		fmt.Printf("%s\n", imgSummary.ID)
	}

}

func listImages(client *client.Client) error {
	var filter filters.Args = filters.NewArgs()
	filter.Add("label", "app=work-env")

	imgSummaries, err := client.ImageList(
		context.Background(),
		types.ImageListOptions{Filters: filter})
	if err != nil {
		return fmt.Errorf("Failed to list environments: %v", err)
	}

	for _, imgSummary := range imgSummaries {
		printImage(&imgSummary)
	}

	return nil
}

func printRunningContainers(containers []types.Container) {
	if len(containers) == 0 {
		return
	}

	maxNameLen := 19
	for _, container := range containers {
		for _, name := range container.Names {
			if len(name) > maxNameLen {
				maxNameLen = len(name)
			}
		}
	}
	format := fmt.Sprintf("%%-%ds %%s\n", maxNameLen)

	fmt.Printf(format, "Environment", "Image")
	for _, container := range containers {
		var name string
		if len(container.Names) > 0 {
			name = strings.Join(container.Names, ",")
		} else {
			name = container.ID
		}

		fmt.Printf(format, name, container.Image)
	}
}

func listContainers(client *client.Client) error {
	var filter filters.Args = filters.NewArgs()
	filter.Add("label", "app=work-env")

	containers, err := client.ContainerList(
		context.Background(),
		types.ContainerListOptions{
			All:     true,
			Filters: filter})
	if err != nil {
		return fmt.Errorf("Failed to list containers: %v", err)
	}

	printRunningContainers(containers)

	return nil
}

func main() {
	var CLI struct {
		Build struct {
			Path  string `arg help:"Docker image build path"`
			Image string `arg help:"Name of the environment image"`
			// DockerFile string `arg help:"DockerFile used to create environment" default:"DockerFile"`
		} `cmd help:"Build new environment image <image-name> from a DockerFile in a current directory"`

		Create struct {
			Image string `arg help:"Name of a Docker image used to create environment"`
			Name  string `arg name:"env-name" help:"Name of the new environment"`
			Rm    bool   `help:"Remove environment after session finished"`
		} `cmd help:"Create new environment instance <env-name> from docker image <image> and attach to it. Overwrites existing containers."`

		Attach struct {
			Name string `arg name:"env-name" help:"Environment name (docker container) to attach"`
			Rm   bool   `help:"Remove environment after session finished"`
		} `cmd help:"Start working in environment. Start a container and attach to it."`

		Rm struct {
			Name string `arg name:"env-name" help:"Environment to remove"`
		} `cmd help:"Remove an environment instance"`
		Images struct {
		} `cmd help:"List environment images"`
		Ps struct {
		} `cmd help:"List running environment images"`
	}

	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	ctx := kong.Parse(&CLI)

	switch ctx.Command() {
	case "build <path> <image>":
		err := buildEnvironment(client, CLI.Build.Path, CLI.Build.Image)
		if err != nil {
			fmt.Printf("Failed to build an environment image: %v\n", err)
		}
	case "create <image> <env-name>":
		_, err := createWorkEnv(client, CLI.Create.Image, CLI.Create.Name)
		if err != nil {
			fmt.Printf("Failed to create work environment: %v\n", err)
			return
		}
		err = attachToContainer(client, CLI.Create.Name)
		if err != nil {
			fmt.Printf("Failed to enter to environment: %v\n", err)
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
			fmt.Printf("Failed to attach to environment: %v\n", err)
		}
	case "rm <env-name>":
		err := removeContainer(client, CLI.Rm.Name)
		if err != nil {
			fmt.Printf("Failed to remove container: %v\n", err)
		}
	case "images":
		err := listImages(client)
		if err != nil {
			fmt.Printf("Failed to list images: %v\n", err)
		}

	case "ps":
		err := listContainers(client)
		if err != nil {
			fmt.Printf("Failed to list environment instances: %v\n", err)
		}
	default:
		panic(ctx.Command())
	}

}
