# Claude

## Run remote-control as side-car

This repo's devcontainer already configures a side-car container that runs the `remote-control` server.
See [docker-compose.claude.yml](/.devcontainer/docker-compose.claude.yml) for details.
Merge the docker-compose file on [devcontainer.json](/.devcontainer/devcontainer.json) to enable the side-car container.

For the first run, you need to login using Claude code CLI and enable the remote-control server.
Run following command on the host:
```bash
$ docker run --rm -it \
  -e CLAUDE_CONFIG_DIR=/.claude \
  -v "${HOME}/.claude:/.claude" \
  ghcr.io/lesomnus/claude:2.1 \
  claude auth login

$ docker run --rm -it \
  -e CLAUDE_CONFIG_DIR=/.claude \
  -v "${HOME}/.claude:/.claude" \
  ghcr.io/lesomnus/claude:2.1 \
  claude remote-control
```
