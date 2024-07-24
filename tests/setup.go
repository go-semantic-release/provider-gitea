package tests

import (
	"fmt"
	"testing"

	"github.com/cybercinch/go-semantic-release-provider-gitea/pkg/provider"
	"github.com/stretchr/testify/require"
)

func setup() {
	server = CreateTestServer()
}

func createTestGiteaRepo(t *testing.T) *provider.GiteaRepository {
	assertions := require.New(t)
	repo := &provider.GiteaRepository{}

	err := repo.Init(map[string]string{
		"gitea_host": server.URL,
		"slug":       fmt.Sprintf("%s/%s", giteaUser, giteaRepo),
		"token":      "token",
	})
	assertions.NoError(err)
	return repo
}

func teardown() {
	server.Close()
}
