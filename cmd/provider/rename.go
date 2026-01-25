package provider

import (
	"fmt"

	"github.com/skevetter/devpod/cmd/flags"
	"github.com/skevetter/devpod/pkg/config"
	"github.com/skevetter/devpod/pkg/provider"
	"github.com/skevetter/devpod/pkg/workspace"
	"github.com/skevetter/log"
	"github.com/spf13/cobra"
)

// RenameCmd holds the cmd flags

type RenameCmd struct {
	*flags.GlobalFlags
}

// NewRenameCmd creates a new command

func NewRenameCmd(globalFlags *flags.GlobalFlags) *cobra.Command {

	cmd := &RenameCmd{

		GlobalFlags: globalFlags,
	}

	return &cobra.Command{

		Use: "rename",

		Short: "Rename a provider",

		Long: `Renames a provider and automatically rebinds all workspaces
that are bound to it to use the new provider name.

Example:
  devpod provider rename my-provider my-new-provider
`,

		Args: cobra.ExactArgs(2),

		RunE: func(cobraCmd *cobra.Command, args []string) error {

			return cmd.Run(cobraCmd, args)

		},
	}

}

// Run executes the command

func (cmd *RenameCmd) Run(cobraCmd *cobra.Command, args []string) error {
	log.Default.Info("Renaming provider and rebinding workspaces...")

	devPodConfig, err := config.LoadConfig(cmd.Context, cmd.Provider)
	if err != nil {
		return err
	}

	oldName := args[0]
	newName := args[1]

	// List all workspaces to determine which ones need rebinding
	workspaces, err := workspace.ListLocalWorkspaces(devPodConfig.DefaultContext, false, log.Default)
	if err != nil {
		return fmt.Errorf("listing workspaces: %w", err)
	}

	// Collect workspaces that need to be rebound
	var workspacesToRebind []*provider.Workspace
	for _, ws := range workspaces {
		if ws.Provider.Name == oldName {
			workspacesToRebind = append(workspacesToRebind, ws)
		}
	}

	if len(workspacesToRebind) > 0 {
		log.Default.Infof("Found %d workspace(s) that will be rebound from provider '%s' to '%s'",
			len(workspacesToRebind), oldName, newName)
		for _, ws := range workspacesToRebind {
			log.Default.Infof("- Workspace: %s", ws.ID)
		}
	} else {
		log.Default.Info("No workspaces found that are bound to this provider")
	}

	// First, rename the provider
	err = provider.RenameProvider(devPodConfig.DefaultContext, oldName, newName)
	if err != nil {
		return fmt.Errorf("failed to rename provider: %w", err)
	}

	log.Default.Infof("Provider successfully renamed from '%s' to '%s'", oldName, newName)

	// Then, rebind all affected workspaces
	var rebindErrors []error
	var successfulRebinds []string

	for _, ws := range workspacesToRebind {
		log.Default.Infof("Rebinding workspace %s to provider %s", ws.ID, newName)
		ws.Provider.Name = newName
		err := provider.SaveWorkspaceConfig(ws)
		if err != nil {
			log.Default.Errorf("Failed to rebind workspace %s: %v", ws.ID, err)
			rebindErrors = append(rebindErrors, fmt.Errorf("failed to rebind workspace %s: %w", ws.ID, err))
		} else {
			successfulRebinds = append(successfulRebinds, ws.ID)
		}
	}

	// Report results
	if len(successfulRebinds) > 0 {
		log.Default.Donef("Successfully rebound %d workspace(s): %v", len(successfulRebinds), successfulRebinds)
	}

	if len(rebindErrors) > 0 {
		log.Default.Errorf("Failed to rebind %d workspace(s)", len(rebindErrors))
		for _, err := range rebindErrors {
			log.Default.Error(err.Error())
		}
	}

	// Return aggregated error if any rebinding failed
	if len(rebindErrors) > 0 {
		if len(rebindErrors) == 1 {
			return rebindErrors[0]
		}
		// Aggregate multiple errors
		errorMsg := fmt.Sprintf("failed to rebind %d workspace(s): ", len(rebindErrors))
		for i, err := range rebindErrors {
			if i > 0 {
				errorMsg += "; "
			}
			errorMsg += err.Error()
		}
		return fmt.Errorf(errorMsg)
	}

	log.Default.Donef("Successfully renamed provider '%s' to '%s' and rebound all associated workspaces", oldName, newName)
	return nil
}
