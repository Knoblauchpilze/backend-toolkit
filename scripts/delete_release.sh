#!/bin/bash

if [ $# -lt 1 ]; then
  echo "Usage: ./create_release.sh v1.2.3"
  exit 1
fi

VERSION=${1}
if ! [[ $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Version format is incorrect. Expected format: vX.Y.Z"
  echo "Usage: ./create_release.sh v1.2.3"
  exit 1
fi

echo "Deleting release ${VERSION}"

# https://stackoverflow.com/questions/5480258/how-can-i-delete-a-remote-tag
git push --delete origin ${VERSION}
git tag --delete ${VERSION}
