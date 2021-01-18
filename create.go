/* `create` command implementation */

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

func mounts() []mount.Mount {
	return []mount.Mount{
		{Type: "bind", Source: "/home", Target: "/home"},
		{Type: "bind", Source: "/dev", Target: "/dev"},
		{Type: "bind", Source: "/sys", Target: "/sys"},
		{Type: "bind", Source: "/tmp", Target: "/tmp"},
		{Type: "bind", Source: "/etc/resolv.conf", Target: "/etc/resolf.conf"},
	}
}

func envVars(image, name string) []string {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return []string{
		fmt.Sprintf("WORK_ENV_IMAGE=%s", image),
		fmt.Sprintf("WORK_ENV_NAME=%s", name),
		fmt.Sprintf("WORK_ENV_USER_SHELL=%s", os.Getenv("SHELL")),
		fmt.Sprintf("WORK_ENV_USER_ID=%s", user.Uid),
		fmt.Sprintf("WORK_ENV_USER_NAME=%s", user.Username),
		fmt.Sprintf("WORK_ENV_USER_PASSWORD=%s", user.Username),
		fmt.Sprintf("DISPLAY=%s", os.Getenv("DISPLAY")),
	}
}

func createWorkEnv(client *client.Client, image, name string) (containerId string, err error) {
	workDir, _ := os.Getwd()

	var containerConf = container.Config{
		Hostname:     name,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
		Env:          envVars(image, name),
		Cmd:          nil, // TODO             strslice.StrSlice   // Command to run when starting the container
		Image:        image,
		Volumes:      nil,
		WorkingDir:   workDir,
		Entrypoint:   nil, //      strslice.StrSlice   // Entrypoint to run when starting the container
		Labels:       map[string]string{"app": WORK_ENV_APP_VAL},
	}

	var hostConf = container.HostConfig{
		Mounts: mounts(),
		NetworkMode: "host",
	}

	resp, err := client.ContainerCreate(
		context.Background(),
		&containerConf,
		&hostConf,
		nil, // networkingConfig
		// nil, // platformConfig
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
