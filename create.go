/* `create` command implementation */

package main

import (
	"context"
	"fmt"
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

func userName() string {
	user, err := user.Current()
	userName := ""
	if err == nil {
		userName = user.Username
	} else {
		fmt.Printf("Failed to get current user: %v", err)
	}

	return userName
}

func envVars(image, name string) []string {
	return []string{
		fmt.Sprintf("WORK_ENV_IMAGE=%s", image),
		fmt.Sprintf("WORK_ENV_NAME=%s", name),
	}
}

func createWorkEnv(client *client.Client, image, name string) (containerId string, err error) {
	workDir, _ := os.Getwd()

	var containerConf = container.Config{
		Hostname:     name,
		User:         userName(),
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
		Labels:       map[string]string{"app": "work-env"},
	}

	var hostConf = container.HostConfig{
		Mounts: mounts(),
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
