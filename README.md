## `github-perm-auditor` checks user account permissions on Github

### Installation
`go get go.mozilla.org/github-perm-auditor`

### Running
`github-perm-auditor` -> follow prompts

### Configuration
`github-perm-auditor` uses env for config.
The following env variables can be exported to configure `github-perm-auditor`

`github-perm-auditor` will prompt you for these when necessary.

- `PERM_DEBUG=false`
- `PERM_GITHUBTOKEN=aGithubAccessToken`
- `PERM_GITHUBUSERNAME=yourGithubUsername`
- `PERM_GITHUBPASSWORD=yourGithubPassword`
