// Copyright 2019-present Vic Shóstak. All rights reserved.
// Use of this source code is governed by Apache 2.0 license
// that can be found in the LICENSE file.

package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/create-go-app/cli/pkg/cgapp"
	"github.com/create-go-app/cli/pkg/registry"
	"github.com/spf13/cobra"
)

// createCmd represents the `create` command.
var createCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create a new project via interactive UI or configuration file",
	Long:    "\nCreate a new project via interactive UI or configuration file.",
	Run:     runCreateCmd,
}

// runCreateCmd represents runner for the `create` command.
var runCreateCmd = func(cmd *cobra.Command, args []string) {
	// Start message.
	cgapp.SendMsg(true, "* * *", "Create a new project via Create Go App CLI v"+registry.CLIVersion+"...", "yellow", true)

	// If config is set and correct, skip survey and use it.
	if useConfigFile && projectConfig != nil {
		// Re-define variables from config file (default is $PWD/.cgapp.yml).
		backend = strings.ToLower(projectConfig["backend"].(string))
		frontend = strings.ToLower(projectConfig["frontend"].(string))
		webserver = strings.ToLower(projectConfig["webserver"].(string))

		// Check, if config file contains `roles` section
		if rolesConfig != nil {
			installAnsibleRoles = true
		}
	} else {
		// Start survey.
		if err := survey.Ask(
			registry.CreateQuestions, &createAnswers, survey.WithIcons(surveyIconsConfig),
		); err != nil {
			cgapp.SendMsg(true, "[ERROR]", err.Error(), "red", true)
			os.Exit(1)
		}

		// If something went wrong, cancel and exit.
		if !createAnswers.AgreeCreation {
			cgapp.SendMsg(true, "[!]", "You're stopped creation of a new project.", "red", false)
			cgapp.SendMsg(false, "[!]", "Run `cgapp create` once again!", "red", true)
			os.Exit(1)
		}

		// Insert empty line.
		cgapp.SendMsg(false, "", "", "", false)

		// Define variables for better display.
		backend = strings.ToLower(createAnswers.Backend)
		frontend = strings.ToLower(createAnswers.Frontend)
		webserver = strings.ToLower(createAnswers.Webserver)
		installAnsibleRoles = createAnswers.InstallAnsibleRoles
	}

	// Start timer.
	startTimer := time.Now()

	// Get current directory.
	currentDir, err := os.Getwd()
	if err != nil {
		cgapp.SendMsg(true, "[ERROR]", err.Error(), "red", true)
		os.Exit(1)
	}

	// Create config files for your project.
	cgapp.SendMsg(false, "*", "Create config files for your project...", "cyan", true)

	// Create configuration files.
	filesToMake := map[string][]byte{
		".gitignore":     registry.EmbedGitIgnore,
		".gitattributes": registry.EmbedGitAttributes,
		".editorconfig":  registry.EmbedEditorConfig,
		"Makefile":       registry.EmbedMakefile,
	}
	if err := cgapp.MakeFiles(currentDir, filesToMake); err != nil {
		cgapp.SendMsg(true, "[ERROR]", err.Error(), "red", true)
		os.Exit(1)
	}

	// Create Ansible playbook with tasks, if not skipped.
	if installAnsibleRoles {
		cgapp.SendMsg(true, "*", "Create Ansible playbook with tasks...", "cyan", true)

		// Create playbook.
		fileToMake := map[string][]byte{
			"deploy-playbook.yml": registry.EmbedDeployPlaybook,
		}
		if err := cgapp.MakeFiles(currentDir, fileToMake); err != nil {
			cgapp.SendMsg(true, "[ERROR]", err.Error(), "red", true)
			os.Exit(1)
		}
	}

	// Create backend files.
	cgapp.SendMsg(true, "*", "Create project backend...", "cyan", true)
	if err := cgapp.CreateProjectFromRegistry(
		&registry.Project{
			Type:       "backend",
			Name:       backend,
			RootFolder: currentDir,
		},
		registry.Repositories,
		registry.RegexpBackendPattern,
	); err != nil {
		cgapp.SendMsg(true, "[ERROR]", err.Error(), "red", true)
		os.Exit(1)
	}

	if frontend != "none" {
		// Create frontend files.
		cgapp.SendMsg(true, "*", "Create project frontend...", "cyan", false)
		if err := cgapp.CreateProjectFromCmd(
			&registry.Project{
				Type:       "frontend",
				Name:       frontend,
				RootFolder: currentDir,
			},
			registry.Commands,
			registry.RegexpFrontendPattern,
		); err != nil {
			cgapp.SendMsg(true, "[ERROR]", err.Error(), "red", true)
			os.Exit(1)
		}
	}

	// Docker containers.
	if webserver != "none" {

		cgapp.SendMsg(true, "* * *", "Configuring Docker containers...", "yellow", false)

		if webserver != "none" {
			// Create container with a web/proxy server.
			cgapp.SendMsg(true, "*", "Create container with web/proxy server...", "cyan", true)
			if err := cgapp.CreateProjectFromRegistry(
				&registry.Project{
					Type:       "webserver",
					Name:       webserver,
					RootFolder: currentDir,
				},
				registry.Repositories,
				registry.RegexpWebServerPattern,
			); err != nil {
				cgapp.SendMsg(true, "[ERROR]", err.Error(), "red", true)
				os.Exit(1)
			}
		}
	}

	// Stop timer
	stopTimer := time.Since(startTimer).String()

	// End message.
	cgapp.SendMsg(true, "* * *", "Completed in "+stopTimer+"!", "yellow", true)
	cgapp.SendMsg(false, "(i)", "A helpful documentation and next steps -> https://create-go.app/", "green", false)
	cgapp.SendMsg(false, "(i)", "Run `cgapp deploy` to deploy your project to a remote server or run on localhost.", "green", true)
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().BoolVarP(
		&useConfigFile,
		"use-config", "c", false,
		"use config file to create a new project or deploy to a remote server (by default, in $PWD/.cgapp.yml)",
	)
}
