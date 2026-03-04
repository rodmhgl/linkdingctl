#!/bin/sh
set -e

# Build a temporary binary to generate shell completions.
# This script is called by goreleaser's before.hooks, before the
# release builds happen, so we compile a throwaway copy here.

rm -rf completions
mkdir -p completions

go build -o /tmp/linkdingctl-completions ./cmd/linkdingctl

/tmp/linkdingctl-completions completion bash > completions/linkdingctl.bash
/tmp/linkdingctl-completions completion zsh  > completions/linkdingctl.zsh
/tmp/linkdingctl-completions completion fish > completions/linkdingctl.fish

rm -f /tmp/linkdingctl-completions
