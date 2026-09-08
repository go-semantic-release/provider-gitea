package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-semantic-release/semantic-release/v2/pkg/provider"
	"github.com/go-semantic-release/semantic-release/v2/pkg/semrel"
	"github.com/stretchr/testify/require"
)

func TestNewGiteaRepository(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)

	var repo *GiteaRepository
	repo = &GiteaRepository{}
	err := repo.Init(map[string]string{})
	assertions.EqualError(err, "gitea host is not set")

	repo = &GiteaRepository{}

	err = repo.Init(map[string]string{
		"gitea_host": server.URL,
		"slug":       fmt.Sprintf("%s/%s", giteaUser, giteaRepo),
		"token":      "token",
	})
	assertions.NoError(err)
}

func TestVersionAndNameReturn(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)
	repo := createTestGiteaRepo(t)
	assertions.Equal("dev", repo.Version())
	assertions.Equal("Gitea", repo.Name())
}

func TestGiteaGetInfo(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)
	repo := createTestGiteaRepo(t)

	repoInfo, err := repo.GetInfo()

	assertions.NoError(err)
	assertions.Equal(giteaDefaultBranch, repoInfo.DefaultBranch)
	assertions.True(repoInfo.Private)
	assertions.Equal(giteaUser, repoInfo.Owner)
	assertions.Equal(giteaRepo, repoInfo.Repo)
}

func TestGiteaGetCommits(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)
	repo := createTestGiteaRepo(t)

	commits, err := repo.GetCommits("", "sa213445t6")

	assertions.NoError(err)
	for i, c := range commits {
		assertions.Equal(c.SHA, giteaCommits[i].SHA)
		assertions.Equal(c.RawMessage, giteaCommits[i].RepoCommit.Message)
		assertions.Equal(c.Annotations["author_name"], giteaCommits[i].Author.FullName)
		assertions.Equal(c.Annotations["author_email"], giteaCommits[i].Author.Email)
		assertions.Equal(c.Annotations["committer_name"], giteaCommits[i].Committer.FullName)
		assertions.Equal(c.Annotations["committer_email"], giteaCommits[i].Committer.Email)
		assertions.Equal(c.Annotations["author_date"], giteaCommits[i].RepoCommit.Author.Date)
		assertions.Equal(c.Annotations["committer_date"], giteaCommits[i].RepoCommit.Committer.Date)
	}
}

func TestGiteaGetCommitsUnknownUser(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)
	repo := &GiteaRepository{}

	err := repo.Init(map[string]string{
		"gitea_host": server.URL,
		"slug":       fmt.Sprintf("%s/%s", giteaUser, giteaRepoNoUser),
		"token":      "token",
	})
	assertions.NoError(err)

	commits, err := repo.GetCommits("", "sa213445t6")

	assertions.NoError(err)
	assertions.Len(commits, 4)

	// Both author and committer resolve to a Gitea user.
	k := commits[0]
	assertions.Equal("1111111111111111111111111111111111111111", k.SHA)
	assertions.Equal("alice", k.Annotations["author_login"])
	assertions.Equal("Alice", k.Annotations["author_name"])
	assertions.Equal("alice@example.com", k.Annotations["author_email"])
	assertions.Equal("alice", k.Annotations["committer_login"])
	assertions.Equal("Alice", k.Annotations["committer_name"])

	// Author email is not registered with a Gitea user: fall back to the
	// identity on the commit and leave the login empty.
	u := commits[1]
	assertions.Equal("2222222222222222222222222222222222222222", u.SHA)
	assertions.Equal("fix: use example command\n", u.RawMessage)
	assertions.Equal("", u.Annotations["author_login"])
	assertions.Equal("Unknown Author", u.Annotations["author_name"])
	assertions.Equal("unknown@example.com", u.Annotations["author_email"])
	assertions.Equal("2026-04-21T16:00:11-04:00", u.Annotations["author_date"])
	assertions.Equal("bob", u.Annotations["committer_login"])
	assertions.Equal("Bob", u.Annotations["committer_name"])
	assertions.Equal("bob@example.com", u.Annotations["committer_email"])
	assertions.Equal("2026-04-21T22:57:30+02:00", u.Annotations["committer_date"])

	// Committer email is not registered with a Gitea user.
	c := commits[2]
	assertions.Equal("3333333333333333333333333333333333333333", c.SHA)
	assertions.Equal("carol", c.Annotations["author_login"])
	assertions.Equal("Carol", c.Annotations["author_name"])
	assertions.Equal("", c.Annotations["committer_login"])
	assertions.Equal("Unregistered Committer", c.Annotations["committer_name"])
	assertions.Equal("unregistered@example.com", c.Annotations["committer_email"])
	assertions.Equal("2026-05-02T09:31:00Z", c.Annotations["committer_date"])

	// Neither author nor committer resolves to a Gitea user.
	n := commits[3]
	assertions.Equal("4444444444444444444444444444444444444444", n.SHA)
	assertions.Equal("feat: add offline mode\n", n.RawMessage)
	assertions.Equal("", n.Annotations["author_login"])
	assertions.Equal("Eve Nobody", n.Annotations["author_name"])
	assertions.Equal("eve@nobody.example", n.Annotations["author_email"])
	assertions.Equal("2026-06-11T14:00:00Z", n.Annotations["author_date"])
	assertions.Equal("", n.Annotations["committer_login"])
	assertions.Equal("Frank Nobody", n.Annotations["committer_name"])
	assertions.Equal("frank@nobody.example", n.Annotations["committer_email"])
	assertions.Equal("2026-06-11T14:05:00Z", n.Annotations["committer_date"])
}

func TestGiteaGetReleases(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)
	repo := createTestGiteaRepo(t)

	testCases := []struct {
		vRange          string
		re              string
		expectedSHA     string
		expectedVersion string
	}{
		{"", "", testSHA, "2020.4.19"},
		{"", "^v[0-9]*", testSHA, "2.0.0"},
		{"2-beta", "", testSHA, "2.1.0-beta"},
		{"3-beta", "", testSHA, "3.0.0-beta.2"},
		{"4-beta", "", testSHA, "4.0.0-beta"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("VersionRange: %s, Regex: %s", tc.vRange, tc.re), func(t *testing.T) {
			releases, err := repo.GetReleases(tc.re)
			assertions.NoError(err)
			release, err := semrel.GetLatestReleaseFromReleases(releases, tc.vRange)
			require.NoError(t, err)
			assertions.Equal(tc.expectedSHA, release.SHA)
			assertions.Equal(tc.expectedVersion, release.Version)
		})
	}
}

func TestGiteaCreateRelease(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)
	repo := createTestGiteaRepo(t)

	err := repo.CreateRelease(&provider.CreateReleaseConfig{
		NewVersion: "5.0.0",
		Prerelease: false,
		Branch:     "",
		SHA:        testSHA,
	})
	assertions.NoError(err)
}

func TestGiteaCreateReleaseStripPrefix(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)
	repo := &GiteaRepository{}

	err := repo.Init(map[string]string{
		"gitea_host":         server.URL,
		"slug":               fmt.Sprintf("%s/%s", giteaUser, giteaRepo),
		"token":              "token",
		"strip_v_tag_prefix": "true",
	})

	assertions.NoError(err)

	err = repo.CreateRelease(&provider.CreateReleaseConfig{
		NewVersion: "5.0.0",
		Prerelease: false,
		Branch:     "",
		SHA:        testSHA,
	})
	assertions.NoError(err)
}

func TestGiteaInvalidTag(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)
	repo := &GiteaRepository{}

	err := repo.Init(map[string]string{
		"gitea_host": server.URL,
		"slug":       fmt.Sprintf("%s/%s", giteaUser, giteaRepo),
		"token":      "token",
	})

	assertions.NoError(err)

	err = repo.CreateRelease(&provider.CreateReleaseConfig{
		NewVersion: "1.0.1",
		Prerelease: false,
		Branch:     "",
		SHA:        testSHA,
	})
	assertions.Errorf(err, "invalid tag name")
}

func TestGiteaEnvironmentVars(t *testing.T) {
	setup()
	defer teardown()

	testCases := []struct {
		name        string
		envVarName  string
		envVarValue string
	}{
		{
			"Github Environment Var Slug",
			"GITHUB_REPOSITORY",
			fmt.Sprintf("%s/%s",
				giteaUser,
				giteaRepo),
		},
		{
			"Gitea Environment Var Slug",
			"GITEA_REPOSITORY",
			fmt.Sprintf("%s/%s",
				giteaUser,
				giteaRepo),
		},
		{
			"WoodpeckerCI Environment Var Slug",
			"CI_REPO_NAME",
			fmt.Sprintf("%s/%s",
				giteaUser,
				giteaRepo),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Setenv(tc.envVarName, tc.envVarValue)

			repo := &GiteaRepository{}
			err := repo.Init(map[string]string{
				"gitea_host": server.URL,
				"token":      "token",
			})

			require.NoError(t, err)
			_ = os.Unsetenv(tc.envVarName)
		})
	}
}

func TestGiteaTokenNotSet(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)

	repo := &GiteaRepository{}
	err := repo.Init(map[string]string{
		"gitea_host": server.URL,
	})

	assertions.Errorf(err, "gitea token missing")
}

func TestGiteaNonBooleanStripPrefix(t *testing.T) {
	setup()
	defer teardown()

	assertions := require.New(t)

	repo := &GiteaRepository{}
	err := repo.Init(map[string]string{
		"gitea_host":         server.URL,
		"slug":               fmt.Sprintf("%s/%s", giteaUser, giteaRepo),
		"strip_v_tag_prefix": "something",
		"token":              "token",
	})

	assertions.Errorf(err, "failed to set property strip_v_tag_prefix: strconv.ParseBool: parsing \"something\": invalid syntax")
}
