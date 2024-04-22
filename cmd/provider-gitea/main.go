package main

import (
	giteaProvider "github.com/cybercinch/go-semantic-release-provider-gitea/pkg/provider"
	"github.com/go-semantic-release/semantic-release/v2/pkg/plugin"
	"github.com/go-semantic-release/semantic-release/v2/pkg/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Provider: func() provider.Provider {
			return &giteaProvider.GiteaRepository{}
		},
	})
}
