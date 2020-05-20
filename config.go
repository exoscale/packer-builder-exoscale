//go:generate mapstructure-to-hcl2 -type Config

package exoscale

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	pkrconfig "github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const (
	defaultAPIEndpoint                  = "https://api.exoscale.com/v1"
	defaultInstanceType                 = "Medium"
	defaultInstanceDiskSize       int64 = 50
	defaultInstanceSecurityGroup        = "default"
	defaultInstanceTemplateFilter       = "featured"
)

type Config struct {
	ctx interpolate.Context

	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	APIEndpoint             string `mapstructure:"api_endpoint"`
	APIKey                  string `mapstructure:"api_key"`
	APISecret               string `mapstructure:"api_secret"`
	InstanceName            string `mapstructure:"instance_name"`
	InstanceZone            string `mapstructure:"instance_zone"`
	InstanceTemplate        string `mapstructure:"instance_template"`
	InstanceTemplateFilter  string `mapstructure:"instance_template_filter"`
	InstanceType            string `mapstructure:"instance_type"`
	InstanceDiskSize        int64  `mapstructure:"instance_disk_size"`
	InstanceSecurityGroup   string `mapstructure:"instance_security_group"`
	InstanceSSHKey          string `mapstructure:"instance_ssh_key"`
	TemplateZone            string `mapstructure:"template_zone"`
	TemplateName            string `mapstructure:"template_name"`
	TemplateDescription     string `mapstructure:"template_description"`
	TemplateUsername        string `mapstructure:"template_username"`
	TemplateDisablePassword bool   `mapstructure:"template_disable_password"`
	TemplateDisableSSHKey   bool   `mapstructure:"template_disable_sshkey"`
}

func NewConfig(raws ...interface{}) (*Config, error) {
	var config = Config{
		APIEndpoint:            defaultAPIEndpoint,
		InstanceType:           defaultInstanceType,
		InstanceDiskSize:       defaultInstanceDiskSize,
		InstanceSecurityGroup:  defaultInstanceSecurityGroup,
		InstanceTemplateFilter: defaultInstanceTemplateFilter,
	}

	err := pkrconfig.Decode(
		&config,
		&pkrconfig.DecodeOpts{
			Interpolate:        true,
			InterpolateContext: &config.ctx,
			InterpolateFilter: &interpolate.RenderFilter{
				Exclude: []string{"run_command"},
			},
		},
		raws...)
	if err != nil {
		return nil, err
	}

	requiredArgs := map[string]interface{}{
		"api_key":           config.APIKey,
		"api_secret":        config.APISecret,
		"api_endpoint":      config.APIEndpoint,
		"instance_template": config.InstanceTemplate,
		"template_zone":     config.TemplateZone,
		"template_name":     config.TemplateName,
	}

	errs := new(packer.MultiError)
	for k, v := range requiredArgs {
		if reflect.ValueOf(v).IsZero() {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("%s must be set", k))
		}
	}

	if config.InstanceZone == "" {
		config.InstanceZone = config.TemplateZone
	}

	if es := config.Comm.Prepare(&config.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if len(errs.Errors) > 0 {
		return nil, errs
	}

	return &config, nil
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }
