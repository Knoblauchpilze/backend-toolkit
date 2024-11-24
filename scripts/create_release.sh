#!/bin/bash

if [[ $# -ge 1 ]]; then
  echo "Using version from input: ${1}"
  VERSION=${1}
fi

if [[ $VERSION == "" ]]; then
  # Try to pick the latest version:
  # https://stackoverflow.com/questions/6269927/how-can-i-list-all-tags-in-my-git-repository-by-the-date-they-were-created
  VERSION=$(git tag --sort=-creatordate | tail -n 1)

  if [[ $VERSION == "" ]]; then
    echo "No versions defined yet, picking 0.0.0 as base revision"
    VERSION="0.0.0"
  fi

  # Add a revision as per:
  # https://en.wikipedia.org/wiki/Software_versioning#Semantic_versioning
  BASE_VERSION=$(echo $VERSION | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+')
  REVISION=$(echo $VERSION | cut -d. -f4)
  if [[ $REVISION == "" ]]; then
    REVISION="0"
  fi

  echo "BASE: $BASE_VERSION, rev: $REVISION"

  NEXT_REVISION=$(echo "${REVISION} + 1" | bc)
  VERSION="${BASE_VERSION}.${NEXT_REVISION}"

  echo "BASE: $BASE_VERSION, rev: $REVISION, next: $NEXT_REVISION"
  echo "No version provided in input, will go on with ${VERSION}"
fi

echo "Creating release ${VERSION}"
# https://stackoverflow.com/questions/18216991/create-a-tag-in-a-github-repository
git tag ${VERSION}

git push origin ${VERSION}
