#!/usr/bin/env bash
set -euo pipefail

BINARY="tinker"
REPO="github.com/mvaliolahi/tinker/cmd/tinker"

info()  { printf "\033[1;34m→\033[0m %s\n" "$1"; }
ok()    { printf "\033[1;32m✓\033[0m %s\n" "$1"; }
err()   { printf "\033[1;31m✗\033[0m %s\n" "$1" >&2; exit 1; }

# --- check go ---
if ! command -v go &>/dev/null; then
	err "Go is not installed — https://go.dev/dl/"
fi

info "Installing ${BINARY}..."

go install "${REPO}@latest"

# --- find binary ---
GOBIN="$(go env GOBIN 2>/dev/null || true)"
GOPATH="$(go env GOPATH 2>/dev/null || true)"
BINDIR="${GOBIN:-${GOPATH:-$HOME/go}/bin}"

if [ ! -f "${BINDIR}/${BINARY}" ]; then
	err "Binary not found at ${BINDIR}/${BINARY}"
fi

ok "Installed ${BINARY} to ${BINDIR}/${BINARY}"

# --- PATH setup ---
case ":${PATH}:" in
	*":${BINDIR}:"*)
		ok "${BINDIR} is already in PATH"
		;;
	*)
		SHELL_RC=""
		case "${SHELL:-}" in
			*/zsh)  SHELL_RC="$HOME/.zshrc" ;;
			*/bash) SHELL_RC="$HOME/.bashrc" ;;
			*/fish) SHELL_RC="$HOME/.config/fish/config.fish" ;;
		esac

		if [ -n "${SHELL_RC}" ]; then
			if [ -f "${SHELL_RC}" ] && grep -q "${BINDIR}" "${SHELL_RC}" 2>/dev/null; then
				ok "${BINDIR} already in ${SHELL_RC}"
			else
				info "Adding ${BINDIR} to ${SHELL_RC}"
				if [[ "${SHELL_RC}" == *fish* ]]; then
					echo "set -gx PATH ${BINDIR} \$PATH" >> "${SHELL_RC}"
				else
					echo "export PATH=\"${BINDIR}:\$PATH\"" >> "${SHELL_RC}"
				fi
				ok "Added to ${SHELL_RC} — restart your shell or run: source ${SHELL_RC}"
			fi
		else
			info "Add to your shell config manually: export PATH=\"${BINDIR}:\$PATH\""
		fi
		;;
esac

# --- verify ---
info "Verifying..."
if command -v "${BINARY}" &>/dev/null; then
	ok "$(${BINARY} version)"
else
	info "Run this to use now: export PATH=\"${BINDIR}:\$PATH\""
fi
