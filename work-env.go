package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	// for docker attach "github.com/moby/moby/pkg/stdcopy"
)

func createWorkEnv(client *client.Client, image, name string) (containerId string, err error) {
	workDir, _ := os.Getwd()

	user, err := user.Current()
	userName := "";
	if err != nil {
		userName = user.Username
	}

	var containerConf = container.Config{
		Hostname: name,
		User: userName,
		AttachStdin: true,
		AttachStdout: true,
		AttachStderr: true,
		Tty: true,
		OpenStdin: true,
		Env: nil, // TODO
		Cmd: nil, // TODO             strslice.StrSlice   // Command to run when starting the container
		Image: image,
		Volumes: map[string]struct{} {
			"/home:/home": {},
			"/dev:/dev": {},
			"/sys:/sys": {},
			"/tmp:/tmp": {}},
		WorkingDir: workDir,
		Entrypoint: nil, //      strslice.StrSlice   // Entrypoint to run when starting the container
		Labels: map[string]string{"app": "work-env"},
	}

	resp, err := client.ContainerCreate(
		context.Background(),
		&containerConf,
		nil, // HostConfig
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

	err := command.Run()
	if err != nil {
		return fmt.Errorf("Failed to execute 'docker exec': %v", err)
	}

	return nil
}


func removeContainer(client *client.Client, containerName string) error {
	return client.ContainerRemove(
		context.Background(),
		containerName,
		types.ContainerRemoveOptions{Force: true})
}

func main() {
	// TODO more sophisticated parser
	var command string = os.Args[1]
	var containerName string = os.Args[2]

	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	switch command {
	case "create":
		image := os.Args[3]  // TODO args parser
		containerId, err := createWorkEnv(client, image, containerName)
		if err != nil {
			fmt.Printf("Failed to create work envinronment: %v\n", err)
		}
		fmt.Printf("Created container %s\n", containerId)
	case "attach":
		err := attachToContainer(client, containerName)
		if err != nil {
			fmt.Printf("Failed to attach to envinronment: %v\n", err)
		}
	case "remove", "rm":
		err := removeContainer(client, containerName)
		if err != nil {
			fmt.Printf("Failed to remove container: %v\n", err)
		}
	}
}
