# :octocat: provider-gitea
[![CI](https://github.com/guisea/go-semantic-release-provider-gitea/actions/workflows/ci.yml/badge.svg)](https://github.com/guisea/go-semantic-release-provider-gitea/actions/workflows/ci.yml)[![codecov](https://codecov.io/github/guisea/go-semantic-release-provider-gitea/graph/badge.svg?token=K8DX5ZQ3Y2)](https://codecov.io/github/guisea/go-semantic-release-provider-gitea)

The Gitea provider for [go-semantic-release](https://github.com/go-semantic-release/semantic-release).

### Provider Option

The provider options can be configured via the `--provider-opt` CLI flag.

| Name | Description                                               | Example                                                             |
|---|-----------------------------------------------------------|---------------------------------------------------------------------|
| gitea_host | This configures the provider to use a Gitea host endpoint | `--provider-opt gitea_host=gitea.example.corp`                      |
| slug | The owner and repository name                             | `--provider-opt slug=cybercinch/go-semantic-release-provider-gitea` |
| token | Gitea Personal Access Token                               | `--provider-opt token=xxx`                                          |
| strip_v_prefix | Strip "v" from release prefix default: false              | `--provider-opt strip_v_tag_prefix=true`                             |

## Licence

The [MIT License (MIT)](http://opensource.org/licenses/MIT)

Copyright © 2024 [Aaron Guise](https://github.com/guisea)
