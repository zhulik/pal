#!/usr/bin/env bash
# Cloud Agent install script for the pal repository.
#
# Idempotently installs the pinned toolchain (task, golangci-lint) that is not
# part of the base image, then warms the Go module caches and generates the
# mocks the test suites depend on. Go itself is provided by the base image and
# pinned via go.mod / .tool-versions (currently 1.26.4).
set -euo pipefail

# Keep these in sync with .tool-versions.
TASK_VERSION="v3.49.1"
GOLANGCI_LINT_VERSION="v2.12.2"
BIN_DIR="/usr/local/bin"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# sudo is only needed when BIN_DIR is not writable by the current user.
maybe_sudo() {
  if [ -w "$BIN_DIR" ]; then
    "$@"
  else
    sudo "$@"
  fi
}

install_task() {
  if command -v task >/dev/null 2>&1 && task --version 2>/dev/null | grep -q "${TASK_VERSION#v}"; then
    echo "task ${TASK_VERSION} already installed"
    return
  fi
  echo "Installing task ${TASK_VERSION}"
  curl -sSfL https://taskfile.dev/install.sh | maybe_sudo sh -s -- -d -b "$BIN_DIR" "$TASK_VERSION"
}

install_golangci_lint() {
  if command -v golangci-lint >/dev/null 2>&1 && golangci-lint --version 2>/dev/null | grep -q "${GOLANGCI_LINT_VERSION#v}"; then
    echo "golangci-lint ${GOLANGCI_LINT_VERSION} already installed"
    return
  fi
  echo "Installing golangci-lint ${GOLANGCI_LINT_VERSION}"
  # Use the install script pinned to the release tag so its embedded checksums match.
  curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/${GOLANGCI_LINT_VERSION}/install.sh" \
    | maybe_sudo sh -s -- -b "$BIN_DIR" "$GOLANGCI_LINT_VERSION"
}

install_task
install_golangci_lint

echo "Downloading Go modules (root)"
go mod download

echo "Downloading Go modules (cmd/pal)"
(cd cmd/pal && go mod download)

# Generate the mocks that the library test suite imports (gitignored, so they
# must exist before `go test` / `golangci-lint` can type-check the tests).
echo "Generating mocks"
go generate ./...

echo "Install complete: $(go version); task $(task --version); $(golangci-lint --version)"
