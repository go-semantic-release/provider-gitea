package provider

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"

	"github.com/stretchr/testify/require"
)

func createGiteaCommit(sha, message, date string) *gitea.Commit {
	tDate, _ := time.Parse("2006-01-02T15:04:05±hh:mm", date)
	return &gitea.Commit{
		CommitMeta: &gitea.CommitMeta{
			URL:     "",
			SHA:     sha,
			Created: tDate,
		},
		HTMLURL: "",
		RepoCommit: &gitea.RepoCommit{
			URL: "",
			Author: &gitea.CommitUser{
				Identity: gitea.Identity{
					Name:  testUserName,
					Email: testUserEmail,
				},
				Date: date,
			},
			Committer: &gitea.CommitUser{
				Identity: gitea.Identity{
					Name:  testUserName,
					Email: testUserEmail,
				},
				Date: date,
			},
			Message: message,
			Tree: &gitea.CommitMeta{
				URL:     "",
				SHA:     "",
				Created: tDate,
			},
			Verification: &gitea.PayloadCommitVerification{
				Verified:  false,
				Reason:    "",
				Signature: "",
				Payload:   "",
			},
		},
		Author: &gitea.User{
			ID:       0,
			UserName: "owner",
			FullName: testUserName,
			Email:    testUserEmail,
		},
		Committer: &gitea.User{
			ID:       0,
			UserName: "owner",
			FullName: testUserName,
			Email:    testUserEmail,
		},
		Parents: nil,
		Files:   nil,
		Stats: &gitea.CommitStats{
			Total:     0,
			Additions: 0,
			Deletions: 0,
		},
	}
}

func retrieveData(filepath string) ([]byte, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)

	byteValue, _ := io.ReadAll(jsonFile)
	return byteValue, nil
}

func setup() {
	server = CreateTestServer()
}

func createTestGiteaRepo(t *testing.T) *GiteaRepository {
	assertions := require.New(t)
	repo := &GiteaRepository{}

	err := repo.Init(map[string]string{
		optKeyHost:  server.URL,
		optKeySlug:  fmt.Sprintf("%s/%s", giteaUser, giteaRepo),
		optKeyToken: testTokenValue,
	})
	assertions.NoError(err)
	return repo
}

func teardown() {
	server.Close()
}
