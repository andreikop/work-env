/* Command line parsing */

package main

import (
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/daemon/names"
)

func formatError(command string, err error) error {
	if err != nil {
		return fmt.Errorf("Failed to %s: %v", command, err)
	} else {
		return nil
	}
}

func validateImageName(name string) error {
	if !reference.ReferenceRegexp.MatchString(name) {
		return fmt.Errorf("Incorrect docker image name '%s'", name)
	}
	return nil
}

func validateContainerName(name string) error {
	if !names.RestrictedNamePattern.MatchString(name) {
		return fmt.Errorf("Incorrect docker container name '%s'", name)
	}
	return nil
}

type BuildCmd struct {
	Path  string `arg help:"Docker image build path"`
	Image string `arg help:"Name of the environment image"`
	// DockerFile string `arg help:"DockerFile used to create environment" default:"DockerFile"`
}

func (b *BuildCmd) Run(ctx *Context) error {
	err := validateImageName(b.Image)
	if err != nil {
		return err
	}

	return formatError(
		"build",
		buildEnvironmentCommand(ctx.client, b.Path, b.Image))
}

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
	err := validateImageName(r.Image)
	if err != nil {
		return err
	}
	err = validateContainerName(r.Name)
	if err != nil {
		return err
	}

	return formatError("run environment",
		runEnvironmentCommand(ctx.client, r.Image, r.Name, r.Overwrite, r.Rm))
}

type PsCmd struct{}

func (p *PsCmd) Run(ctx *Context) error {
	return formatError("list environment instances",
		listContainersCommand(ctx.client))
}

type EnterCmd struct {
	Name string `arg name:"env-name" default:"work-env" help:"Environment name (docker container) to attach"`
}

func (e *EnterCmd) Run(ctx *Context) error {
	err := validateContainerName(e.Name)
	if err != nil {
		return err
	}

	return formatError("Enter container",
		enterContainerCommand(ctx.client, e.Name))
}

type RmCmd struct {
	Names []string `arg name:"env-name" help:"Environment to remove"`
}

func (r *RmCmd) Run(ctx *Context) error {
	for _, name := range r.Names {
		err := validateContainerName(name)
		if err != nil {
			return err
		}
	}

	return formatError("remove container", removeContainerCommand(ctx.client, r.Names))
}

type RmImageCmd struct {
	Images []string `arg name:"image-name" help:"Docker image name to remove. Only images built by work-env can be removed."`
}

func (r *RmImageCmd) Run(ctx *Context) error {
	for _, imageName := range r.Images {
		err := validateImageName(imageName)
		if err != nil {
			return err
		}
	}

	return formatError("remove image", removeImageCommand(ctx.client, r.Images))
}

var CLI struct {
	Build  BuildCmd   `cmd help:"Build new environment image <image-name> from a DockerFile in a current directory"`
	Images ImagesCmd  `cmd help:"List environment images"`
	Run    RunCmd     `cmd help:"Create a new environment instance <env-name> from docker image <image> and attach to it. Overwrites existing containers."`
	Ps     PsCmd      `cmd help:"List running environment images"`
	Enter  EnterCmd   `cmd help:"Start working in an environment instance. Start a container if not running and attach to it."`
	Rm     RmCmd      `cmd help:"Remove an environment instance"`
	Rmi    RmImageCmd `cmd help:"Remove a docker image built by work-env"`
}
