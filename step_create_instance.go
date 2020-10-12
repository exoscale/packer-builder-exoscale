package exoscale

import (
	"context"
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateInstance struct{}

func (s *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		buildID = state.Get("build-id").(string)
		exo     = state.Get("exo").(*egoscale.Client)
		ui      = state.Get("ui").(packer.Ui)
		config  = state.Get("config").(*Config)
		zone    = state.Get("zone").(*egoscale.Zone)
	)

	ui.Say("Creating Compute instance")

	instanceName := config.InstanceName
	if instanceName == "" {
		instanceName = "packer-" + buildID
	}

	resp, err := exo.GetWithContext(ctx, &egoscale.ListServiceOfferings{Name: config.InstanceType})
	if err != nil {
		ui.Error(fmt.Sprintf("unable to list Compute instance types: %s", err))
		return multistep.ActionHalt
	}
	instanceType := resp.(*egoscale.ServiceOffering)

	resp, err = exo.GetWithContext(ctx, &egoscale.ListTemplates{
		Name:           config.InstanceTemplate,
		TemplateFilter: config.InstanceTemplateFilter,
		ZoneID:         zone.ID,
	})
	if err != nil {
		ui.Error(fmt.Sprintf("unable to list Compute instance templates: %s", err))
		return multistep.ActionHalt
	}
	instanceTemplate := resp.(*egoscale.Template)

	// If not set at this point, attempt to retrieve the template's username to set the SSH communicator's username.
	if config.Comm.SSHUsername == "" {
		if username, ok := instanceTemplate.Details["username"]; ok {
			config.Comm.SSHUsername = username
		}
	}

	privateNetworks, err := instancePrivateNetworkIDs(ctx, state)
	if err != nil {
		ui.Error(fmt.Sprintf("unable to retrieve Private Networks: %s", err))
		return multistep.ActionHalt
	}

	resp, err = exo.RequestWithContext(ctx, &egoscale.DeployVirtualMachine{
		Name:               instanceName,
		ServiceOfferingID:  instanceType.ID,
		TemplateID:         instanceTemplate.ID,
		RootDiskSize:       config.InstanceDiskSize,
		KeyPair:            config.InstanceSSHKey,
		SecurityGroupNames: []string{config.InstanceSecurityGroup},
		NetworkIDs:         privateNetworks,
		ZoneID:             zone.ID,
	})
	if err != nil {
		ui.Error(fmt.Sprintf("unable to create Compute instance: %s", err))
		return multistep.ActionHalt
	}
	instance := resp.(*egoscale.VirtualMachine)
	state.Put("instance", instance)
	state.Put("instance_ip_address", instance.IP().String())

	if config.PackerDebug {
		ui.Message(fmt.Sprintf("Compute instance started (ID: %s)", instance.ID.String()))
	}

	return multistep.ActionContinue
}

func (s *stepCreateInstance) Cleanup(_ multistep.StateBag) {}

func instancePrivateNetworkIDs(ctx context.Context, state multistep.StateBag) ([]egoscale.UUID, error) {
	var (
		exo    = state.Get("exo").(*egoscale.Client)
		config = state.Get("config").(*Config)
		zone   = state.Get("zone").(*egoscale.Zone)
		ids    []egoscale.UUID
	)

	resp, err := exo.RequestWithContext(ctx, &egoscale.ListNetworks{ZoneID: zone.ID})
	if err != nil {
		return nil, err
	}

next:
	for _, userNetwork := range config.InstancePrivateNetworks {
		for _, network := range resp.(*egoscale.ListNetworksResponse).Network {
			if network.Name == userNetwork {
				ids = append(ids, *network.ID)
				continue next
			}
		}

		return nil, fmt.Errorf("%q: not found", userNetwork)
	}

	return ids, nil
}
