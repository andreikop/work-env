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

func checkWorkEnvImageExists(client *client.Client, name string) error {
	imInspect, _, err := client.ImageInspectWithRaw(context.Background(), name)
	if err != nil {
		return err
	}

	appVal, ok := imInspect.Config.Labels["app"]
	if !ok {
		return fmt.Errorf("Image '%s' is not a work-env image. Label 'app' not found", name)
	}
	if appVal != WORK_ENV_APP_VAL {
		return fmt.Errorf("Image '%s' is not a work-env container. Label 'app' equals to %s", name, appVal)
	}

	return nil
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

func enterContainer(client *client.Client, containerName string) error {
	confJson, err := getWorkEnvContainer(client, containerName)
	if err != nil {
		return err
	}

	args := []string{"exec", "-it", containerName, confJson.Path}
	args = append(args, confJson.Args...)
	// Run container entrypoint once more
	command := exec.Command("/usr/bin/docker", args...)

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

func enterContainerCommand(client *client.Client, containerName string) error {
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

	return enterContainer(client, containerName)
}

func removeContainerCommand(client *client.Client, containerNames []string) error {
	for _, containerName := range containerNames {
		_, err := getWorkEnvContainer(client, containerName)
		if err != nil {
			return err
		}
	}

	for _, containerName := range containerNames {
		err := client.ContainerRemove(
			context.Background(),
			containerName,
			types.ContainerRemoveOptions{Force: true})
		if err != nil {
			return err
		}
	}

	return nil
}

func removeImageCommand(client *client.Client, imageNames []string) error {
	for _, imageName := range imageNames {
		err := checkWorkEnvImageExists(client, imageName)
		if err != nil {
			return err
		}
	}

	for _, imageName := range imageNames {
		_, err := client.ImageRemove(
			context.Background(),
			imageName,
			types.ImageRemoveOptions{Force: true, PruneChildren: true})

		if err != nil {
			return err
		}
	}

	return nil
}

func printImage(imgSummary *types.ImageSummary) {
	if len(imgSummary.RepoTags) > 0 {
		for _, repoTag := range imgSummary.RepoTags {
			if strings.Count(repoTag, ":") == 1 { // name:version
				parts := strings.Split(repoTag, ":")
				name := parts[0]
				version := parts[1]
				if name == "<none>" { // strange names :-/
					continue
				}

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

func attachToContainer(client *client.Client, containerName string) error {
	_, err := getWorkEnvContainer(client, containerName)
	if err != nil {
		return err
	}

	command := exec.Command("/usr/bin/docker", "attach", containerName) // FIXME works only one time

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

func runEnvironmentCommand(client *client.Client, image, name string, overwrite, rmAfter bool) error {
	_, err := getWorkEnvContainer(client, name)
	alreadyExists := err == nil

	if alreadyExists {
		if overwrite {
			err := removeContainerCommand(client, []string{name})
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Container '%s' is already running. Add --overwire to overwire it", name)
		}
	}

	_, err = createWorkEnv(client, image, name)
	if err != nil {
		return fmt.Errorf("Failed to create work environment: %v", err)
	}
	err = attachToContainer(client, name)
	if err != nil {
		return fmt.Errorf("Failed to enter to environment: %v", err)
	}
	if rmAfter {
		err = removeContainerCommand(client, []string{name})
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
	kongObj := kong.Must(&CLI,
		kong.Name("work-env"),
		kong.Description("Virtual command line working environment for developers. http://github.com/andreikop/work-env"))

	ctx, err := kongObj.Parse(os.Args[1:])
	kongObj.FatalIfErrorf(err)

	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	_, err = client.Ping(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Failed to connect to Docker daemon. Error: '%v'.\nIs Docker daemon running and does current user have sufficient access permissions?\n",
			err)
		return
	}

	err = ctx.Run(&Context{client: client})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
