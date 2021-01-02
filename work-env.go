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

// Get container by name. Verify if it is a work-env container
func getWorkEnvContainer(client *client.Client, name string) (*types.ContainerJSON, error) {
	json, err := client.ContainerInspect(context.Background(), name)

	if err != nil {
		return nil, err
	}

	appVal, ok := json.Config.Labels["app"]
	if !ok {
		return nil, fmt.Errorf("Container '%s' is not a work-env container. Label 'app' not found", name)
	}
	if appVal != WORK_ENV_APP_VAL {
		return nil, fmt.Errorf("Container '%s' is not a work-env container. Label 'app' equals to %s", name, appVal)
	}

	return &json, nil
}

func buildEnvironmentCommand(client *client.Client, path, image string) error {
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
	_, err := getWorkEnvContainer(client, containerName)
	if err != nil {
		return err
	}

	command := exec.Command("/usr/bin/docker", "exec", "-it", containerName, "zsh") // FIXME zsh hardcoded

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

func attachToContainerCommand(client *client.Client, containerName string) error {
	contJson, err := getWorkEnvContainer(client, containerName)
	if err != nil {
		return err
	}
	if !contJson.State.Running {
		err = client.ContainerStart(context.Background(), containerName, types.ContainerStartOptions{})
		if err != nil {
			return fmt.Errorf("Failed to start container: %s", err)
		}
	}

	return attachToContainer(client, containerName)
}

func removeContainerCommand(client *client.Client, containerName string) error {
	_, err := getWorkEnvContainer(client, containerName)
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

func listImagesCommand(client *client.Client) error {
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

func runEnvironmentCommand(client *client.Client, image, name string, rm bool) error {
	// TODO remove running container
	_, err := createWorkEnv(client, image, name)
	if err != nil {
		return fmt.Errorf("Failed to create work environment: %v", err)
	}
	err = attachToContainer(client, name)
	if err != nil {
		return fmt.Errorf("Failed to enter to environment: %v", err)
	}
	if rm {
		err = removeContainerCommand(client, name)
		if err != nil {
			return fmt.Errorf("Failed to remove container: %v", err)
		}
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

func listContainersCommand(client *client.Client) error {
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

type Context struct {
	client *client.Client
}

func main() {
	var CLI struct {
		Build  BuildCmd  `cmd help:"Build new environment image <image-name> from a DockerFile in a current directory"`
		Images ImagesCmd `cmd help:"List environment images"`
		Run    RunCmd    `cmd help:"Create a new environment instance <env-name> from docker image <image> and attach to it. Overwrites existing containers."`
		Ps     PsCmd     `cmd help:"List running environment images"`
		Attach AttachCmd `cmd help:"Start working in an environment instance. Start a container and attach to it."`
		Rm     RmCmd     `cmd help:"Remove an environment instance"`
		// TODO rmi command
	}

	ctx := kong.Parse(&CLI)

	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// TODO check if docker daemon is alive
	err = ctx.Run(&Context{client: client})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
