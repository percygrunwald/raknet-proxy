#!/usr/bin/env bash

set -e

# Install all the asdf plugins and then run asdf install
# Assumes asdf is installed and in the path
# shellcheck disable=SC1091
. "$ASDF_DIR/asdf.sh"
INSTALLED_PLUGINS=$(asdf plugin-list || echo "")
while read PLUGIN _VERSION; do
  if echo $INSTALLED_PLUGINS | grep -w $PLUGIN > /dev/null; then
    printf "Plugin %s already installed.\n" $PLUGIN
  else
    printf "Installing plugin %s...\n" $PLUGIN
    asdf plugin add $PLUGIN
  fi
done <.tool-versions
asdf install

# Install required go tools at the system level
while read GO_DEP; do
  printf "Installing go package %s...\n" $GO_DEP
  go install $GO_DEP
done <.go-deps
asdf reshim golang
