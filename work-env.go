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

	"github.com/alecthomas/kong"
)

const WORK_ENV_APP_VAL = "work-env"

func verifyWorkEnvContainer(client *client.Client, name string) error {
	json, err := client.ContainerInspect(context.Background(), name)

	if err != nil {
		return err
	}

	appVal, ok := json.Config.Labels["app"]
	if !ok {
		return fmt.Errorf("Container '%s' is not a work-env container. Label 'app' not found", name)
	}
	if appVal != WORK_ENV_APP_VAL {
		return fmt.Errorf("Container '%s' is not a work-env container. Label 'app' equals to %s", name, appVal)
	}

	return nil
}

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
	err := verifyWorkEnvContainer(client, containerName)
	if err != nil {
		return err
	}

	command := exec.Command("/usr/bin/docker", "exec", "-it", containerName, "zsh")

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Start()
	if err != nil {
		return fmt.Errorf("Failed to start 'docker exec': %v", err)
	}
	defer command.Wait()

	return nil
}

func removeContainer(client *client.Client, containerName string) error {
	err := verifyWorkEnvContainer(client, containerName)
	if err != nil {
		return err
	}

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

func imgAndContainerFilter() filters.Args {
	var filter filters.Args = filters.NewArgs()
	filter.Add("label", fmt.Sprintf("app=%s", WORK_ENV_APP_VAL))
	return filter
}

func listImages(client *client.Client) error {
	imgSummaries, err := client.ImageList(
		context.Background(),
		types.ImageListOptions{Filters: imgAndContainerFilter()})
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
		if strings.HasPrefix(name, "/") { // Don't know why docker appends / to names
			name = name[1:]
		}

		fmt.Printf(format, name, container.Image)
	}
}

func listContainers(client *client.Client) error {
	containers, err := client.ContainerList(
		context.Background(),
		types.ContainerListOptions{
			All:     true,
			Filters: imgAndContainerFilter()})
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
		Images struct {
		} `cmd help:"List environment images"`
		Run struct {
			Image string `arg help:"Name of a Docker image used to create an environment"`
			Name  string `arg name:"env-name" help:"Name of the new environment"`
			Rm    bool   `help:"Remove an environment after a session finished"`
		} `cmd help:"Create a new environment instance <env-name> from docker image <image> and attach to it. Overwrites existing containers."`
		Ps struct {
		} `cmd help:"List running environment images"`
		Attach struct {
			Name string `arg name:"env-name" help:"Environment name (docker container) to attach"`
			Rm   bool   `help:"Remove environment after session finished"`
		} `cmd help:"Start working in an environment instance. Start a container and attach to it."`
		Rm struct {
			Name string `arg name:"env-name" help:"Environment to remove"`
		} `cmd help:"Remove an environment instance"`
	}

	ctx := kong.Parse(&CLI)

	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	switch ctx.Command() {
	case "build <path> <image>":
		err := buildEnvironment(client, CLI.Build.Path, CLI.Build.Image)
		if err != nil {
			fmt.Printf("Failed to build an environment image: %v\n", err)
		}
	case "run <image> <env-name>":
		// TODO autogenerate env-name
		// TODO remove running container
		_, err := createWorkEnv(client, CLI.Run.Image, CLI.Run.Name)
		if err != nil {
			fmt.Printf("Failed to create work environment: %v\n", err)
			return
		}
		err = attachToContainer(client, CLI.Run.Name)
		if err != nil {
			fmt.Printf("Failed to enter to environment: %v\n", err)
			return
		}
		if CLI.Run.Rm {
			err = removeContainer(client, CLI.Run.Name)
			if err != nil {
				fmt.Printf("Failed to remove container: %v\n", err)
			}
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
	case "attach <env-name>":
		// FIXME ensure container is running
		err := attachToContainer(client, CLI.Attach.Name)
		if err != nil {
			fmt.Printf("Failed to attach to a container: %v\n", err)
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
