# :octocat: provider-gitea
[![CI](https://github.com/cybercinch/go-semantic-release-provider-gitea/workflows/CI/badge.svg?branch=master)](https://github.com/cybercinch/go-semantic-release-provider-gitea/actions?query=workflow%3ACI+branch%3Amaster)

The Gitea provider for [go-semantic-release](https://github.com/go-semantic-release/semantic-release).

### Provider Option

The provider options can be configured via the `--provider-opt` CLI flag.

| Name | Description | Example |
|---|---|---|
| gitea_host | This configures the provider to use a Gitea host endpoint | `--provider-opt gitea_host=gitea.example.corp` |
| slug | The owner and repository name  | `--provider-opt slug=cybercinch/go-semantic-release-provider-gitea` |
| token | Gitea Personal Access Token  | `--provider-opt token=xxx` |

## Licence

The [MIT License (MIT)](http://opensource.org/licenses/MIT)

Copyright © 2024 [Aaron Guise](https://github.com/guisea)
