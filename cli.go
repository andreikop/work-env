/* Command line parsing */

package main

import (
	"fmt"
)

func formatError(command string, err error) error {
	if err != nil {
		return fmt.Errorf("Failed to %s: %v", command, err)
	} else {
		return nil
	}
}

type BuildCmd struct {
	Path  string `arg help:"Docker image build path"`
	Image string `arg help:"Name of the environment image"`
	// DockerFile string `arg help:"DockerFile used to create environment" default:"DockerFile"`
}

func (b *BuildCmd) Run(ctx *Context) error {
	return formatError(
		"build",
		buildEnvironmentCommand(ctx.client, b.Path, b.Image))
}

// TODO validate image and container names
// https://github.com/docker/distribution/blob/master/reference/regexp.go
type ImagesCmd struct {
}

func (i *ImagesCmd) Run(ctx *Context) error {
	return formatError("list images",
		listImagesCommand(ctx.client))
}

type RunCmd struct {
	Image     string `arg default:"work-env" help:"Name of a Docker image used to create an environment"`
	Name      string `arg name:"env-name" default:"work-env" help:"Name of the new environment"`
	Overwrite bool   `short:"y" help:"Overwrite existing container if exists"`
	Rm        bool   `help:"Remove an environment after a session finished"`
}

func (r *RunCmd) Run(ctx *Context) error {
	return formatError("run environment",
		runEnvironmentCommand(ctx.client, r.Image, r.Name, r.Overwrite, r.Rm))
}

type PsCmd struct{}

func (p *PsCmd) Run(ctx *Context) error {
	return formatError("list environment instances",
		listContainersCommand(ctx.client))
}

type AttachCmd struct {
	Name string `arg name:"env-name" default:"work-env" help:"Environment name (docker container) to attach"`
	Rm   bool   `help:"Remove environment after session finished"`
}

func (a *AttachCmd) Run(ctx *Context) error {
	return formatError("attach to a container",
		attachToContainerCommand(ctx.client, a.Name))
}

type RmCmd struct {
	Name string `arg name:"env-name" help:"Environment to remove"`
}

func (r *RmCmd) Run(ctx *Context) error {
	return formatError("remove container", removeContainerCommand(ctx.client, r.Name))
}
