package tests

import (
	"io"
	"os"
	"time"

	"code.gitea.io/sdk/gitea"
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
					Name:  "A User",
					Email: "owner@noreply.1.1.1.1",
				},
				Date: date,
			},
			Committer: &gitea.CommitUser{
				Identity: gitea.Identity{
					Name:  "A User",
					Email: "owner@noreply.1.1.1.1",
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
			FullName: "A User",
			Email:    "owner@noreply.1.1.1.1",
		},
		Committer: &gitea.User{
			ID:       0,
			UserName: "owner",
			FullName: "A User",
			Email:    "owner@noreply.1.1.1.1",
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
