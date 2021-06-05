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
				err := validateImageName(image)
				if err != nil {
					return err
				}
			}

			var name = "work-env"
			if len(args) > 1 {
				name = args[1]
				err := validateContainerName(name)
				if err != nil {
					return err
				}
			}

			return formatError("run environment",
				runEnvironmentCommand(
					globalClient,
					image, name,
					cmd.LocalFlags().Lookup("overwrite").Value.String() == "true",
					cmd.LocalFlags().Lookup("remove").Value.String() == "true"))
		},
	}

	cmdPs = &cobra.Command{
		Use: "ps",
		Short: "List environment instances",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return formatError(
				"list environment instances",
				listContainersCommand(globalClient))
		},
	}

	cmdEnter = &cobra.Command{
		Use: "enter [env-name=work-env]",
		Short: "Start working in an environment instance. Start a container <env-name> if not running and attach to it.",
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var envName = "work-env"
			if len(args) > 0 {
				envName = args[0]
				err := validateContainerName(envName)
				if err != nil {
					return err
				}
			}

			return formatError("Enter container",
				enterContainerCommand(globalClient, envName))
		},
	}

	// FIXME document 1+ args
	cmdRm = &cobra.Command{
		Use: "rm env-name",
		Short: "Remove an environment instance <env-name>",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range(args) {
				err := validateContainerName(arg)
				if err != nil {
					return err
				}
			}

			return formatError("Remove container",
				removeContainerCommand(globalClient, args))
		},
	}

	// FIXME document 1+ args
	cmdRmi = &cobra.Command{
		Use: "rmi image-name",
		Short: "Remove a docker image <image-env> built by work-env",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range(args) {
				err := validateImageName(arg)
				if err != nil {
					return err
				}
			}

			return formatError("Remove image",
				removeImageCommand(globalClient, args))
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

	rootCmd.AddCommand(cmdPs)
	rootCmd.AddCommand(cmdEnter)
	rootCmd.AddCommand(cmdRm)
	rootCmd.AddCommand(cmdRmi)

	rootCmd.Execute()
}
