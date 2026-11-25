#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
rm -rf data && mkdir -p data/content
cp ../projects.json data/
cp ../content/*.md data/content/
