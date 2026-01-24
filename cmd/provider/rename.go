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

		Long: `

		#######################################################

		############### devpod provider rename #################

		#######################################################

		Renames a provider.



		WARNING: Renaming a provider will cause all workspaces bound to it to fail. You will have to manually edit the workspace configurations to use the new provider name.



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
	log.Default.Warn("Renaming a provider might break existing workspaces, please make sure to check them after the provider is renamed.")

	devPodConfig, err := config.LoadConfig(cmd.Context, cmd.Provider)
	if err != nil {
		return err
	}

	oldName := args[0]
	newName := args[1]

	workspaces, err := workspace.ListLocalWorkspaces(devPodConfig.DefaultContext, false, log.Default)
	if err != nil {
		return fmt.Errorf("listing workspaces: %w", err)
	}

	for _, ws := range workspaces {
		if ws.Provider.Name == oldName {
			log.Default.Infof("Rebinding workspace %s to provider %s", ws.ID, newName)
			ws.Provider.Name = newName
			err := provider.SaveWorkspaceConfig(ws)
			if err != nil {
				log.Default.Warnf("Failed to rebind workspace %s: %v", ws.ID, err)
			}
		}
	}

	err = provider.RenameProvider(devPodConfig.DefaultContext, oldName, newName)
	if err != nil {
		return fmt.Errorf("failed to rename provider: %w", err)
	}

	return nil

}
