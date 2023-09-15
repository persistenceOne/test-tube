#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PERSISTENCECORE_REV=${1:-main}

LATEST_PERSISTENCECORE_VERSION="v16"

# if "$OSMOIS_REV" is /v\d+/ then extract it as var
if [[ "$PERSISTENCECORE_REV" =~ ^v[0-9]+ ]]; then
  PERSISTENCECORE_VERSION=$(echo "$PERSISTENCECORE_REV" | sed "s/\..*$//")
else
  PERSISTENCECORE_VERSION="$LATEST_PERSISTENCECORE_VERSION"
fi

########################################
## Update and rebuild persistence-test-tube ##
########################################

# update all submodules
git submodule update --init --recursive
cd "$SCRIPT_DIR/../packages/persistence-test-tube/persistencecore"

PERSISTENCECORE_REV_NO_ORIGIN="$(echo "$PERSISTENCECORE_REV" | sed "s/^origin\///")"

git checkout "$PERSISTENCECORE_REV_NO_ORIGIN"


# build and run update-persistence-test-tube
cd "$SCRIPT_DIR/update-persistencecore-test-tube-deps" && go build

# run update-persistence-test-tube-deps which will replace the `replace directives` in persistence-test-tube
# with persistencecore' replaces
"$SCRIPT_DIR/update-persistence-test-tube-deps/update-persistence-test-tube-deps" "$PERSISTENCECORE_REV_NO_ORIGIN"

cd "$SCRIPT_DIR/../packages/persistence-test-tube/persistencecore"
PARSED_REV=$(git rev-parse --short "$PERSISTENCECORE_REV")

cd "$SCRIPT_DIR/../packages/persistence-test-tube/libpersistencetesttube"

go get "github.com/persistenceOne/persistenceCore/${PERSISTENCECORE_VERSION}@${PARSED_REV}"

# tidy up updated go.mod
go mod tidy


########################################
## Update git revision if there is    ##
## any change                         ##
########################################

if [[ -n "${SKIP_GIT_UPDATE:-}" ]]; then
  echo '[SKIP] SKIP_GIT_UPDATE is set, skipping git update'
  exit 0
fi

# if dirty or untracked file exists
if [[ $(git diff --stat) != '' ||  $(git ls-files  --exclude-standard  --others) ]]; then
  # add, commit and push
  git add "$SCRIPT_DIR/.."
  git commit -m "rebuild with $(git rev-parse --short HEAD:dependencies/persistencecore)"

  # remove "origin/"
  PERSISTENCECORE_REV=$(echo "$PERSISTENCECORE_REV" | sed "s/^origin\///")
  BRANCH="autobuild-$PERSISTENCECORE_REV"

  # force delete local "$BRANCH" if exists
  git branch -D "$BRANCH" || true

  git checkout -b "$BRANCH"
  git push -uf origin "$BRANCH"
else
  echo '[CLEAN] No update needed for this build'
fi
