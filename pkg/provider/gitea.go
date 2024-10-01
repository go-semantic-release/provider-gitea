package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/Masterminds/semver/v3"
	"github.com/go-semantic-release/semantic-release/v2/pkg/provider"
	"github.com/go-semantic-release/semantic-release/v2/pkg/semrel"
)

var PVERSION = "dev"

type GiteaRepository struct {
	client          *gitea.Client
	repo            string
	owner           string
	stripVTagPrefix bool
	baseURL         string
}

// gocyclo:ignore
func (repo *GiteaRepository) Init(config map[string]string) error {
	giteaHost := config["gitea_host"]
	if giteaHost == "" {
		giteaHost = os.Getenv("GITEA_HOST")
	}
	// If host is still not set error
	if giteaHost == "" {
		return fmt.Errorf("gitea host is not set")
	}

	repo.baseURL = giteaHost
	slug := config["slug"]

	if slug == "" {
		slug = os.Getenv("GITHUB_REPOSITORY")
	}
	// Maybe we are running in Gitea Actions
	if slug == "" {
		slug = os.Getenv("GITEA_REPOSITORY")
	}
	// Maybe we are running in WoodpeckerCI
	if slug == "" {
		slug = os.Getenv("CI_REPO_NAME")
	}

	token := config["token"]
	if token == "" {
		token = os.Getenv("GITEA_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("gitea token missing")
	}

	if !strings.Contains(slug, "/") {
		return fmt.Errorf("invalid slug")
	}
	split := strings.Split(slug, "/")
	// This could be due to act locally
	// We'll work backwards to get the values
	repo.owner = split[len(split)-2]
	repo.repo = split[len(split)-1]

	// Ensure no .git suffix remains
	repo.repo = strings.TrimSuffix(repo.repo, ".git")

	ctx := context.Background()

	client, err := gitea.NewClient(giteaHost,
		gitea.SetToken(token),
		gitea.SetContext(ctx))
	if err != nil {
		return err
	}

	repo.client = client

	stripVTagPrefix := config["strip_v_tag_prefix"]
	repo.stripVTagPrefix, err = strconv.ParseBool(stripVTagPrefix)

	if stripVTagPrefix != "" && err != nil {
		return fmt.Errorf("failed to set property strip_v_tag_prefix: %w", err)
	}

	return nil
}

func (repo *GiteaRepository) GetInfo() (*provider.RepositoryInfo, error) {
	r, _, err := repo.client.GetRepo(repo.owner, repo.repo)
	if err != nil {
		return nil, err
	}
	return &provider.RepositoryInfo{
		Owner:         r.Owner.UserName,
		Repo:          r.Name,
		DefaultBranch: r.DefaultBranch,
		Private:       r.Private,
	}, nil
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

			if commit.Author == nil {
				return nil, fmt.Errorf("gitea: author is not found. Check email [%s] is assigned to user",
					commit.RepoCommit.Author.Email)
			}

			if commit.Committer == nil {
				return nil, fmt.Errorf("gitea: committer is not found. Check email [%s] is assigned to user",
					commit.RepoCommit.Committer.Email)
			}

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

//gocyclo:ignore
func (repo *GiteaRepository) GetReleases(rawRe string) ([]*semrel.Release, error) {
	re := regexp.MustCompile(rawRe)
	allReleases := make([]*semrel.Release, 0)
	opts := gitea.ListRepoTagsOptions{ListOptions: gitea.ListOptions{PageSize: 100}}
	for {
		refs, resp, err := repo.client.GetRepoRefs(repo.owner, repo.repo, "")
		if resp != nil && resp.StatusCode == 404 {
			return allReleases, nil
		}
		if err != nil {
			return nil, err
		}
		for _, r := range refs {
			tag := strings.TrimPrefix(r.Ref, "refs/tags/")
			if rawRe != "" && !re.MatchString(tag) {
				continue
			}
			objType := r.Object.Type
			if objType != "commit" && objType != "tag" {
				continue
			}
			foundSha := r.Object.SHA
			// resolve annotated tag
			if objType == "tag" {
				resTag, _, err := repo.client.GetRepoRef(repo.owner, repo.repo, foundSha)
				if err != nil {
					continue
				}
				if resTag.Object.Type != "commit" {
					continue
				}
				foundSha = resTag.Object.SHA
			}
			version, err := semver.NewVersion(tag)
			if err != nil {
				continue
			}
			allReleases = append(allReleases, &semrel.Release{SHA: foundSha, Version: version.String()})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allReleases, nil
}

func (repo *GiteaRepository) CreateRelease(release *provider.CreateReleaseConfig) error {
	prefix := "v"
	if repo.stripVTagPrefix {
		prefix = ""
	}

	tag := prefix + release.NewVersion
	isPrerelease := release.Prerelease || semver.MustParse(release.NewVersion).Prerelease() != ""

	opt := gitea.CreateReleaseOption{
		TagName:      tag,
		Target:       release.Branch,
		Title:        tag,
		Note:         release.Changelog,
		IsPrerelease: isPrerelease,
	}

	_, _, err := repo.client.CreateRelease(repo.owner, repo.repo, opt)
	return err
}

func (repo *GiteaRepository) Name() string {
	return "Gitea"
}

func (repo *GiteaRepository) Version() string {
	return PVERSION
}
