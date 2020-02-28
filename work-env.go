package main

import (
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func createWorkEnv(client *client.Client, image, name string) error {
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
			"/tmp:/tmp": {}},
		WorkingDir: workDir,
		Entrypoint: nil, //      strslice.StrSlice   // Entrypoint to run when starting the container
		Labels: map[string]string{"app": "work-env"},
	}

	_, err = client.ContainerCreate(
		context.Background(),
		&containerConf,
		nil, // HostConfig
		nil, // networkingConfig
		name)

	if err != nil {
		return fmt.Errorf("Failed to start container: %v", err)
	}

	return nil
}


func main() {
	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	image := "docker-env"
	containerName := "env"

	err = createWorkEnv(client, image, containerName)
	if err != nil {
		panic(err)
	}
}
