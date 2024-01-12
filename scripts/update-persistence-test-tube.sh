#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PERSISTENCE_REV=${1:-main}

LATEST_PERSISTENCE_VERSION="v9"

# if "$PERSISTENCE_REV" is /v\d+/ then extract it as var
if [[ "$PERSISTENCE_REV" =~ ^v[0-9]+ ]]; then
  PERSISTENCE_VERSION=$(echo "$PERSISTENCE_REV" | sed "s/\..*$//")
else
  PERSISTENCE_VERSION="$LATEST_PERSISTENCE_VERSION"
fi

##############################################
## Update and rebuild persistence-test-tube ##
##############################################

# update all submodules
git submodule update --init --recursive
cd "$SCRIPT_DIR/../packages/persistence-test-tube/persistencecore"

PERSISTENCE_REV_NO_ORIGIN="$(echo "$PERSISTENCE_REV" | sed "s/^origin\///")"

git checkout "$PERSISTENCE_REV_NO_ORIGIN"


# build and run update-persistence-test-tube
cd "$SCRIPT_DIR/update-persistence-test-tube-deps" && go build

# run update-persistence-test-tube-deps which will replace the `replace directives` in persistence-test-tube
# with persistence's replaces
"$SCRIPT_DIR/update-persistence-test-tube-deps/update-persistence-test-tube-deps" "$PERSISTENCE_REV_NO_ORIGIN"

cd "$SCRIPT_DIR/../packages/persistence-test-tube/persistencecore"
PARSED_REV=$(git rev-parse --short "$PERSISTENCE_REV")

cd "$SCRIPT_DIR/../packages/persistence-test-tube/libpersistencetesttube"

go get "github.com/persistenceOne/persistenceCore/${PERSISTENCE_VERSION}@${PARSED_REV}"

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
  git commit -m "rebuild with $(git rev-parse --short HEAD:dependencies/persistenceCore)"

  # remove "origin/"
  PERSISTENCE_REV=$(echo "$PERSISTENCE_REV" | sed "s/^origin\///")
  BRANCH="autobuild-$PERSISTENCE_REV"

  # force delete local "$BRANCH" if exists
  git branch -D "$BRANCH" || true

  git checkout -b "$BRANCH"
  git push -uf origin "$BRANCH"
else
  echo '[CLEAN] No update needed for this build'
fi
