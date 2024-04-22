package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/Masterminds/semver/v3"
	GitProvider "github.com/go-semantic-release/provider-git/pkg/provider"
	"github.com/go-semantic-release/semantic-release/v2/pkg/provider"
	"github.com/go-semantic-release/semantic-release/v2/pkg/semrel"
)

var PVERSION = "dev"

type GiteaRepository struct {
	client          *gitea.Client
	localRepo       *GitProvider.Repository
	repo            string
	owner           string
	stripVTagPrefix bool
	compareCommits  bool
	baseURL         string
}

//gocyclo:ignore
func (repo *GiteaRepository) Init(config map[string]string) error {
	giteaHost := config["gitea_host"]
	if giteaHost == "" {
		giteaHost = os.Getenv("GITEA_HOST")
	}
	repo.baseURL = giteaHost
	slug := config["slug"]
	if slug == "" {
		slug = os.Getenv("GITHUB_REPOSITORY")
	}
	// Maybe we are running in WoodpeckerCI
	if slug == "" {
		slug = os.Getenv("CI_REPO_NAME")
	}

	token := config["token"]
	if token == "" {
		token = os.Getenv("GITEA_TOKEN")
		repo.localRepo = &GitProvider.Repository{}
		err := repo.localRepo.Init(map[string]string{
			"remote_name": "origin",
			"git_path":    os.Getenv("CI_PROJECT_DIR"),
		})
		if err != nil {
			return errors.New("failed to initialize local git repository: " + err.Error())
		}
	}
	if token == "" {
		return errors.New("gitea token missing")
	}

	if !strings.Contains(slug, "/") {
		return errors.New("invalid slug")
	}
	split := strings.Split(slug, "/")
	repo.owner = split[0]
	repo.repo = split[1]

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

	if config["github_use_compare_commits"] == "true" {
		repo.compareCommits = true
	}

	var err error
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

func (repo *GiteaRepository) GetCommits(fromSha, toSha string) ([]*semrel.RawCommit, error) {
	return repo.localRepo.GetCommits(fromSha, toSha)
}

//gocyclo:ignore
func (repo *GiteaRepository) GetReleases(rawRe string) ([]*semrel.Release, error) {
	return repo.localRepo.GetReleases(rawRe)
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
		Target:       "main",
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
