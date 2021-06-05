/* Command line parsing */

package main

import (
	"errors"
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/daemon/names"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
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

/*

type PsCmd struct{}

func (p *PsCmd) Run(ctx *Context) error {
	return formatError("list environment instances",
		listContainersCommand(ctx.client))
}

type EnterCmd struct {
	Name string `arg name:"env-name" default:"work-env" help:"Environment name (docker container) to attach. Default name is 'work-env'"`
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
	Run    RunCmd     `cmd help:"Create a new environment instance <env-name> from docker image <image> and attach to it."`
	Ps     PsCmd      `cmd help:"List running environment images"`
	Enter  EnterCmd   `cmd help:"Start working in an environment instance. Start a container if not running and attach to it."`
	Rm     RmCmd      `cmd help:"Remove an environment instance"`
	Rmi    RmImageCmd `cmd help:"Remove a docker image built by work-env"`
}

*/

var (
    globalClient *client.Client = nil

    cmdBuild = &cobra.Command{
		Use:   "build [path] [image-name]",
		Short: "Build from a DockerFile in a directory <path> a new environment image <image-name>",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("requires 2 arguments")
			}
			err := validateImageName(args[1])
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var path = args[0]
			var image = args[1]

			err := validateImageName(image)
			if err != nil {
				return err
			}

			return formatError(
				"build",
				buildEnvironmentCommand(globalClient, path, image))
		},
	}

	cmdImages = &cobra.Command{
		Use: "images",
		Short: "List environment images",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return formatError("list images",
				listImagesCommand(globalClient))
		},
	}

	cmdRun = &cobra.Command{
		Use: "run [env-name=work-env] [image=work-env]",
		Short: "Create a new environment instance <env-name> from docker image <image> and attach to it.",
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var image = "work-env"
			if len(args) > 0 {
				image = args[0]
			}

			var name = "work-env"
			if len(args) > 1 {
				name = args[1]
			}

			err := validateImageName(image)
			if err != nil {
				return err
			}
			err = validateContainerName(name)
			if err != nil {
				return err
			}

			return formatError("run environment",
				runEnvironmentCommand(
					globalClient,
					image, name,
					cmd.LocalFlags().Lookup("overwrite").Value.String() == "true",
					cmd.LocalFlags().Lookup("remove").Value.String() == "true"))
		},
	}

)

func executeCommandLine(client *client.Client) {
	globalClient = client
	var rootCmd = &cobra.Command{
		Use:   "work-env",
		Short: "Virtual command line working environment for developers. http://github.com/andreikop/work-env",
	}

	rootCmd.AddCommand(cmdBuild)
	rootCmd.AddCommand(cmdImages)

	// FIXME remove this flag, invent better behavior
	cmdRun.Flags().BoolP("overwrite", "y", true, "Overwrite existing container if exists")
	cmdRun.Flags().BoolP("remove", "r", false, "Remove an environment after a session finished")
	rootCmd.AddCommand(cmdRun)

	rootCmd.Execute()
}
