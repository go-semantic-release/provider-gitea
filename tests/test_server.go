package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"code.gitea.io/sdk/gitea"
)

var validTags = map[string]bool{
	"v5.0.0": true,
	"5.0.0":  true,
}

var (
	server             *httptest.Server
	giteaUser          = "owner"
	giteaRepo          = "test-repo"
	giteaDefaultBranch = "master"
	giteaCommits       = []*gitea.Commit{
		createGiteaCommit(testSHA, "fix:  Removed lint as not go project\n", "2024-04-23T13:20:33+12:00"),
		createGiteaCommit(testSHA, "fix: Oops,  need a tidyup\n", "2024-04-23T13:17:11+12:00"),
		createGiteaCommit(testSHA, "fix: Update CI\n", "2024-04-23T13:15:00+12:00"),
	}
	testSHA = "deadbeef"
)

func CreateTestServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(GiteaHandler))

	return ts
}

//gocyclo:ignore
func GiteaHandler(w http.ResponseWriter, r *http.Request) {
	// Rate Limit headers
	if r.Header.Get("Authorization") == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet && r.URL.Path == "/api/v1/version" {
		// Client performs a request to check version
		// Get json string from file
		data, _ := retrieveData("data/Version.json")
		_, _ = fmt.Fprint(w, string(data))
		return
	}

	if r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/api/v1/repos/%s/%s", giteaUser, giteaRepo) {
		// Get json string from file
		data, _ := retrieveData("data/GetRepoInfo.json")
		_, _ = fmt.Fprint(w, string(data))
		return
	}

	if r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/api/v1/repos/%s/%s/commits", giteaUser, giteaRepo) {
		// Get json string from file
		data, _ := retrieveData("data/GetCommits.json")
		_, _ = fmt.Fprint(w, string(data))
		return
	}

	if r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/api/v1/repos/%s/%s/git/refs/", giteaUser, giteaRepo) {
		// Get json string from file
		data, _ := retrieveData("data/GetRefs.json")
		_, _ = fmt.Fprint(w, string(data))
		return
	}

	if r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf("/api/v1/repos/%s/%s/releases",
		giteaUser,
		giteaRepo) {
		var data map[string]string
		_ = json.NewDecoder(r.Body).Decode(&data)
		_ = r.Body.Close()

		if _, ok := validTags[data["tag_name"]]; !ok {
			http.Error(w, "invalid tag name", http.StatusBadRequest)
			return
		}

		fmt.Fprint(w, "{}")
		return
	}

	http.Error(w, "invalid route", http.StatusNotImplemented)
}
