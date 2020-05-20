package exoscale

import (
	"context"
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepDeleteSSHKey struct{}

func (s *stepDeleteSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		exo    = state.Get("exo").(*egoscale.Client)
		ui     = state.Get("ui").(packer.Ui)
		config = state.Get("config").(*Config)
	)

	// Check if we actually have to delete the SSH key (i.e. if a throwaway key has been created)
	if _, ok := state.GetOk("delete_ssh_key"); !ok {
		return multistep.ActionContinue
	}

	ui.Say("Deleting SSH key")

	err := exo.BooleanRequestWithContext(ctx, &egoscale.DeleteSSHKeyPair{Name: config.InstanceSSHKey})
	if err != nil {
		ui.Error(fmt.Sprintf("unable to delete SSH key: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDeleteSSHKey) Cleanup(state multistep.StateBag) {}
