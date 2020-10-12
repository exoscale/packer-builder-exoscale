package exoscale

import (
	"context"
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepDestroyInstance struct{}

func (s *stepDestroyInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		exo      = state.Get("exo").(*egoscale.Client)
		ui       = state.Get("ui").(packer.Ui)
		instance = state.Get("instance").(*egoscale.VirtualMachine)
	)

	ui.Say("Destroying Compute instance")

	_, err := exo.RequestWithContext(ctx, &egoscale.DestroyVirtualMachine{ID: instance.ID})
	if err != nil {
		ui.Error(fmt.Sprintf("unable to destroy Compute instance: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDestroyInstance) Cleanup(_ multistep.StateBag) {}
