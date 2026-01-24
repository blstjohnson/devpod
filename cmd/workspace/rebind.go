package workspace

import (
	"fmt"

	"github.com/skevetter/devpod/cmd/flags"
	"github.com/skevetter/devpod/pkg/config"
	"github.com/skevetter/devpod/pkg/provider"
	"github.com/skevetter/log"
	"github.com/spf13/cobra"
)

// RebindCmd holds the cmd flags
type RebindCmd struct {
	*flags.GlobalFlags
}

// NewRebindCmd creates a new command
func NewRebindCmd(globalFlags *flags.GlobalFlags) *cobra.Command {
	cmd := &RebindCmd{
		GlobalFlags: globalFlags,
	}

	return &cobra.Command{
		Use:   "rebind <workspace-name> <new-provider-name>",
		Short: "Rebinds a workspace to a new provider",
		Args:  cobra.ExactArgs(2),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.Run(args)
		},
	}
}

// Run executes the command
func (cmd *RebindCmd) Run(args []string) error {
	workspaceName := args[0]
	newProviderName := args[1]

	devPodConfig, err := config.LoadConfig(cmd.Context, cmd.Provider)
	if err != nil {
		return err
	}

	workspaceConfig, err := provider.LoadWorkspaceConfig(devPodConfig.DefaultContext, workspaceName)
	if err != nil {
		return fmt.Errorf("loading workspace config: %w", err)
	}

	log.Default.Infof("Rebinding workspace %s from provider %s to %s", workspaceName, workspaceConfig.Provider.Name, newProviderName)

	workspaceConfig.Provider.Name = newProviderName

	err = provider.SaveWorkspaceConfig(workspaceConfig)
	if err != nil {
		return fmt.Errorf("saving workspace config: %w", err)
	}

	log.Default.Infof("Workspace %s rebound to provider %s", workspaceName, newProviderName)

	return nil
}
