package main

import (
	"context"
	"flag"
	"log"

	"github.com/atanaspam/terraform-provider-splunkacs/internal/splunkacs"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// TODO: Update this string with the published name of your provider.
		Address: "registry.terraform.io/atanaspam/splunkacs",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), splunkacs.New, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
