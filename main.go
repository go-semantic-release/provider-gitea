package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-semantic-release/semantic-release/v2/pkg/semrel"
)

type GiteaRepository struct {
	client          *gitea.Client
	repo            string
	owner           string
	stripVTagPrefix bool
	baseURL         string
}

//gocyclo:ignore
func (repo *GiteaRepository) Init() error {
	giteaHost := "https://hub.cybercinch.nz"
	repo.baseURL = giteaHost
	slug := "cybercinch/ansible-role-common"
	token := "dff97800f1f8238e3c0f6a45f14f6ba3b477a4f5"
	if !strings.Contains(slug, "/") {
		return errors.New("invalid slug")
	}
	split := strings.Split(slug, "/")
	// This could be due to act locally
	// We'll work backwards to get the values
	repo.owner = split[len(split)-2]
	repo.repo = split[len(split)-1]

	// Ensure no .git suffix remains
	repo.repo = strings.TrimSuffix(repo.repo, ".git")

	ctx := context.Background()
	if giteaHost != "" {
		client, err := gitea.NewClient(giteaHost,
			gitea.SetToken(token),
			gitea.SetContext(ctx))
		if err != nil {
			return err
		}
		repo.client = client
	}

	var err error
	stripVTagPrefix := "false"
	repo.stripVTagPrefix, err = strconv.ParseBool(stripVTagPrefix)

	if stripVTagPrefix != "" && err != nil {
		return fmt.Errorf("failed to set property strip_v_tag_prefix: %w", err)
	}

	return nil
}

func (repo *GiteaRepository) GetCommits(_, toSha string) ([]*semrel.RawCommit, error) {
	allCommits := make([]*semrel.RawCommit, 0)
	opts := &gitea.ListOptions{PageSize: 100}
	done := false
	for {
		commits, resp, err := repo.client.ListRepoCommits(repo.owner, repo.repo, gitea.ListCommitOptions{
			SHA:         toSha,
			ListOptions: *opts,
		})

		if err != nil {
			return nil, err
		}
		for _, commit := range commits {
			sha := commit.SHA
			fmt.Println(spew.Sdump(commit))
			allCommits = append(allCommits, &semrel.RawCommit{
				SHA:        sha,
				RawMessage: commit.RepoCommit.Message,
				Annotations: map[string]string{
					"author_login":    commit.Author.UserName,
					"author_name":     commit.Author.FullName,
					"author_email":    commit.Author.Email,
					"author_date":     commit.RepoCommit.Author.Date,
					"committer_login": commit.Committer.UserName,
					"committer_name":  commit.Committer.FullName,
					"committer_email": commit.Committer.Email,
					"committer_date":  commit.RepoCommit.Committer.Date,
				},
			})
		}
		if done || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allCommits, nil
}

func main() {
	c := &GiteaRepository{}
	c.Init()

	commits, _ := c.GetCommits("0902ffb7682f11a22d85a5532231e68497f53afd", "")

	println(spew.Sdump(commits))

}
