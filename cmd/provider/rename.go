package provider

import (
	"fmt"
	"strings"

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
		Use:   "rename",
		Short: "Rename a provider",
		Long: `Renames a provider by cloning it with the new name, automatically rebinds all workspaces
that are bound to it to use the new provider name, and cleans up the old provider.

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
	log.Default.Info("renaming provider using clone and rebinding workspaces")

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

	// Clone the provider with the new name
	_, cloneErr := workspace.CloneProvider(devPodConfig, newName, oldName, log.Default)
	if cloneErr != nil {
		return fmt.Errorf("failed to clone provider: %w", cloneErr)
	}

	log.Default.Infof("Provider successfully cloned from '%s' to '%s'", oldName, newName)

	// Rebind all affected workspaces to the new provider
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

	// If all rebinding succeeded, try to update default provider if needed
	if len(rebindErrors) == 0 && devPodConfig.Current().DefaultProvider == oldName {
		// Update the default provider to the new name
		devPodConfig.Current().DefaultProvider = newName
		err := config.SaveConfig(devPodConfig)
		if err != nil {
			log.Default.Errorf("Failed to update default provider to '%s': %v", newName, err)
			rebindErrors = append(rebindErrors, fmt.Errorf("failed to update default provider: %w", err))
		} else {
			log.Default.Infof("Updated default provider from '%s' to '%s'", oldName, newName)
		}
	}

	// If any rebinding or default provider update failed, we need to rollback and delete the cloned provider
	if len(rebindErrors) > 0 {
		log.Default.Info("Rebinding or default provider update failed, rolling back changes...")

		// Rollback successful rebinds back to the old provider name
		for _, wsID := range successfulRebinds {
			workspaceConfig, loadErr := provider.LoadWorkspaceConfig(devPodConfig.DefaultContext, wsID)
			if loadErr != nil {
				log.Default.Errorf("Failed to load workspace %s for rollback: %v", wsID, loadErr)
				continue
			}

			log.Default.Infof("Rolling back workspace %s to original provider %s", wsID, oldName)
			workspaceConfig.Provider.Name = oldName
			if rollbackErr := provider.SaveWorkspaceConfig(workspaceConfig); rollbackErr != nil {
				log.Default.Errorf("Failed to rollback workspace %s: %v", wsID, rollbackErr)
			}
		}

		// Delete the cloned provider
		err := DeleteProviderConfig(devPodConfig, newName, true)
		if err != nil {
			log.Default.Errorf("Failed to delete cloned provider %s during cleanup: %v", newName, err)
			return fmt.Errorf("failed to rebind workspaces and failed to cleanup cloned provider: %v; original error: %v", err, rebindErrors[0])
		}

		log.Default.Infof("Cloned provider %s deleted successfully", newName)

		// Return aggregated error if any rebinding failed
		if len(rebindErrors) == 1 {
			return rebindErrors[0]
		}
		// Aggregate multiple errors using strings.Builder for efficiency
		var errorMsg strings.Builder
		fmt.Fprintf(&errorMsg, "failed to rebind %d workspace(s): ", len(rebindErrors))
		for i, err := range rebindErrors {
			if i > 0 {
				errorMsg.WriteString("; ")
			}
			errorMsg.WriteString(err.Error())
		}
		return fmt.Errorf("%s", errorMsg.String())
	}

	// Now delete the old provider
	deleteErr := DeleteProviderConfig(devPodConfig, oldName, true)
	if deleteErr != nil {
		log.Default.Errorf("Failed to delete old provider %s: %v", oldName, deleteErr)
		return fmt.Errorf("failed to delete old provider after successful rename: %w", deleteErr)
	}

	log.Default.Infof("Old provider %s deleted successfully", oldName)

	log.Default.Donef("Successfully renamed provider '%s' to '%s' and rebound all associated workspaces", oldName, newName)
	return nil
}
