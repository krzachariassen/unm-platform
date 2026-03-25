#!/usr/bin/env bash
# publish.sh — Build, version, and push unm-platform to ghcr.io
#
# Usage:
#   ./scripts/publish.sh            # bump patch (default)
#   ./scripts/publish.sh minor      # bump minor
#   ./scripts/publish.sh major      # bump major
#   ./scripts/publish.sh 2.1.0      # set explicit version
#
# Required env vars:
#   GITHUB_USER    your GitHub username
#   GITHUB_TOKEN   a token with write:packages scope
#                  (Settings → Developer settings → Personal access tokens)

set -euo pipefail

# ── Config ───────────────────────────────────────────────────────────────────

IMAGE_NAME="unm-platform"
BUMP="${1:-patch}"

# Resolve GitHub user from env or git remote
GITHUB_USER="${GITHUB_USER:-krzachariassen}"

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "Error: set GITHUB_TOKEN (needs write:packages scope)"
  exit 1
fi

REGISTRY="ghcr.io/${GITHUB_USER}/${IMAGE_NAME}"

# ── Version ──────────────────────────────────────────────────────────────────

CURRENT=$(git tag --sort=-v:refname 2>/dev/null | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1 || echo "")
CURRENT="${CURRENT:-v0.0.0}"
CURRENT="${CURRENT#v}"

IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

if [[ "$BUMP" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  VERSION="$BUMP"
elif [[ "$BUMP" == "major" ]]; then
  VERSION="$((MAJOR + 1)).0.0"
elif [[ "$BUMP" == "minor" ]]; then
  VERSION="${MAJOR}.$((MINOR + 1)).0"
elif [[ "$BUMP" == "patch" ]]; then
  VERSION="${MAJOR}.${MINOR}.$((PATCH + 1))"
else
  echo "Usage: $0 [patch|minor|major|X.Y.Z]"
  exit 1
fi

TAG="v${VERSION}"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Publishing ${IMAGE_NAME} ${TAG}"
echo " Registry:  ${REGISTRY}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ── Login ────────────────────────────────────────────────────────────────────

echo "▶ Logging in to ghcr.io..."
echo "${GITHUB_TOKEN}" | docker login ghcr.io -u "${GITHUB_USER}" --password-stdin

# ── Build ────────────────────────────────────────────────────────────────────

echo "▶ Building image..."
docker build \
  -t "${REGISTRY}:${VERSION}" \
  -t "${REGISTRY}:latest" \
  .

# ── Push ─────────────────────────────────────────────────────────────────────

echo "▶ Pushing ${REGISTRY}:${VERSION}..."
docker push "${REGISTRY}:${VERSION}"

echo "▶ Pushing ${REGISTRY}:latest..."
docker push "${REGISTRY}:latest"

# ── Git tag ──────────────────────────────────────────────────────────────────

if git rev-parse "${TAG}" &>/dev/null; then
  echo "▶ Git tag ${TAG} already exists, skipping"
else
  echo "▶ Tagging commit as ${TAG}..."
  git tag "${TAG}"
  if git remote get-url origin &>/dev/null; then
    git push origin "${TAG}"
  else
    echo "  (no git remote — tag created locally only)"
  fi
fi

# ── Done ─────────────────────────────────────────────────────────────────────

echo ""
echo "✓ Published:"
echo "  ${REGISTRY}:${VERSION}"
echo "  ${REGISTRY}:latest"
echo ""
echo "On your server:"
echo "  export GITHUB_USER=${GITHUB_USER}"
echo "  docker compose -f docker-compose.server.yml pull && docker compose -f docker-compose.server.yml up -d"
echo ""
echo "Note: if the package is private, authenticate on the server first:"
echo "  echo \$GITHUB_TOKEN | docker login ghcr.io -u ${GITHUB_USER} --password-stdin"
