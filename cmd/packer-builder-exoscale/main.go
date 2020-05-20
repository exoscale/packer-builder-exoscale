package main

import (
	"github.com/hashicorp/packer/packer/plugin"

	exoscale "github.com/exoscale/packer-builder-exoscale"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}

	if err := server.RegisterBuilder(new(exoscale.Builder)); err != nil {
		panic(err)
	}

	server.Serve()
}
