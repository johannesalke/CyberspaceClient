#!/bin/sh
# cyberspacecli installer
#
#   curl -fsSL https://raw.githubusercontent.com/johannesalke/cyberspacecli/main/install.sh | sh
#
# Commands (through a pipe, pass them after `sh -s --`):
#   install              install the client (default)
#   update | upgrade     fetch the latest version and reinstall
#   remove | uninstall   delete the installed binary and build directory
#   help                 show usage
#
# Environment overrides:
#   CYBERSPACE_REPO         owner/repo to install from
#   CYBERSPACE_REF          git ref or release tag to install (default: main)
#   CYBERSPACE_CMD          command name to install as (default: cyberspacecli)
#   CYBERSPACE_INSTALL_DIR  binary directory (default: ~/.local/bin)
#   CYBERSPACE_SRC_DIR      source/state directory (default: ~/.local/share/cyberspacecli)
#   CYBERSPACE_FROM_SOURCE=1  always build from source, never download a release

set -eu

REPO="${CYBERSPACE_REPO:-johannesalke/cyberspacecli}"
REF="${CYBERSPACE_REF:-main}"
CMD="${CYBERSPACE_CMD:-cyberspacecli}"
SCRIPT_URL="https://raw.githubusercontent.com/johannesalke/cyberspacecli/main/install.sh"

# Everything lives in main() so a truncated download can never execute a
# half-read script.
main() {
	case "${1:-install}" in
	install) cmd_install install ;;
	update | upgrade) cmd_install update ;;
	remove | uninstall) cmd_remove "${2:-}" ;;
	help | -h | --help) usage ;;
	*)
		printf 'error: unknown command: %s\n\n' "$1" >&2
		usage
		exit 1
		;;
	esac
}

usage() {
	cat >&2 <<EOF
cyberspacecli installer

  install              install the client (default)
  update, upgrade      fetch the latest version and reinstall
  remove, uninstall    delete the installed binary and build directory
                       (pass --purge to also delete $(config_dir))
  help                 show this message

Through a pipe, pass the command after \`sh -s --\`:

  curl -fsSL $SCRIPT_URL | sh
  curl -fsSL $SCRIPT_URL | sh -s -- update
  curl -fsSL $SCRIPT_URL | sh -s -- remove

Environment:
  CYBERSPACE_REPO         owner/repo to install from
  CYBERSPACE_REF          git ref or release tag to install (default: main)
  CYBERSPACE_CMD          command name to install as (default: cyberspacecli)
  CYBERSPACE_INSTALL_DIR  binary directory (default: ~/.local/bin)
  CYBERSPACE_SRC_DIR      source/state directory
  CYBERSPACE_FROM_SOURCE=1  always build from source, never download a release
EOF
}

cmd_install() {
	mode="$1"
	install_dir=$(resolve_install_dir)

	need_cmd mkdir
	need_cmd chmod
	need_cmd mktemp
	need_cmd uname

	if [ "$mode" = update ] && [ -n "$(installed_binary)" ]; then
		say "current: $(installed_ref)"
	elif [ "$mode" = update ]; then
		say "$CMD is not installed — installing it fresh"
	fi

	# A published binary is the cheap path: no Go toolchain, no compile. Source
	# is the fallback, and the only path when installing a branch rather than a
	# release tag.
	if ! try_release "$install_dir"; then
		build_from_source "$install_dir"
	fi

	say "installed $install_dir/$CMD ($(installed_ref))"
	warn_path "$install_dir"
	say "run it with: $CMD"
}

# Delete what this script installed. Idempotent: removing nothing is a success,
# so this is safe to run twice or on a machine that never had it.
cmd_remove() {
	purge="${1:-}"
	removed=0
	src_dir=$(resolve_src_dir)
	cfg_dir=$(config_dir)

	binary=$(installed_binary)
	if [ -n "$binary" ]; then
		# Only ever delete a binary this script recorded installing. Anything
		# else named `cyberspacecli` on the PATH belongs to someone else — a
		# `go install` copy, a package manager's, a build the user made.
		if [ "$binary" = "$(recorded_binary)" ]; then
			remove_path "$binary"
			say "removed $binary"
			removed=1
		else
			say "note: $binary was not installed by this script — leaving it alone"
		fi
	fi

	if [ -d "$src_dir" ]; then
		# CYBERSPACE_SRC_DIR is user input and an `rm -rf` of the wrong
		# directory is unrecoverable, so check it is ours before deleting.
		if is_our_source "$src_dir" || [ -f "$src_dir/installed-at" ]; then
			rm -rf "$src_dir"
			say "removed $src_dir"
			removed=1
		else
			say "note: $src_dir is not a cyberspacecli checkout — leaving it alone"
		fi
	fi

	if [ "$purge" = "--purge" ] && [ -d "$cfg_dir" ]; then
		rm -rf "$cfg_dir"
		say "removed $cfg_dir (saved session and settings)"
		removed=1
	fi

	if [ "$removed" -eq 0 ]; then
		say "$CMD is not installed — nothing to remove"
		return
	fi

	# Config holds the saved refresh token, so it outlives an uninstall: a
	# reinstall should not force a re-login.
	if [ "$purge" != "--purge" ] && [ -d "$cfg_dir" ]; then
		say "kept your config at $cfg_dir — re-run with --purge to delete it"
	fi

	# A second copy earlier on the PATH would make the removal look like it failed.
	remaining=$(installed_binary)
	if [ -n "$remaining" ]; then
		say "note: another copy is still on your PATH at $remaining"
	fi
}

say() { printf '  %s\n' "$*" >&2; }
err() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

have() { command -v "$1" >/dev/null 2>&1; }
need_cmd() { have "$1" || err "required command not found: $1"; }

resolve_src_dir() {
	if [ -n "${CYBERSPACE_SRC_DIR:-}" ]; then
		echo "$CYBERSPACE_SRC_DIR"
	else
		echo "${XDG_DATA_HOME:-$HOME/.local/share}/cyberspacecli"
	fi
}

# A compiled binary would work in a system directory, but defaulting to one
# means sudo on every install. Default to a user directory and let
# CYBERSPACE_INSTALL_DIR override when a shared install is genuinely wanted.
resolve_install_dir() {
	if [ -n "${CYBERSPACE_INSTALL_DIR:-}" ]; then
		echo "$CYBERSPACE_INSTALL_DIR"
	else
		echo "$HOME/.local/bin"
	fi
}

# Where the client itself keeps its config — mirrors GetConfigDir() in
# internal/cyberspaceClient/main_client.go. Only used to tell the user about it
# on uninstall, and to delete it on --purge.
config_dir() {
	case "$(uname -s 2>/dev/null || echo unknown)" in
	Darwin) echo "$HOME/Library/Application Support/cyberspace_client" ;;
	*) echo "${XDG_CONFIG_HOME:-$HOME/.config}/cyberspace_client" ;;
	esac
}

# Print the path of the installed binary, or nothing.
#
# An explicit CYBERSPACE_INSTALL_DIR is a hard boundary: only that directory is
# considered, so `remove` can never reach outside the directory it was pointed at.
installed_binary() {
	if [ -n "${CYBERSPACE_INSTALL_DIR:-}" ]; then
		if [ -f "$CYBERSPACE_INSTALL_DIR/$CMD" ]; then
			echo "$CYBERSPACE_INSTALL_DIR/$CMD"
		fi
		return 0
	fi

	found=$(command -v "$CMD" 2>/dev/null || true)
	if [ -n "$found" ] && [ -f "$found" ]; then
		echo "$found"
		return 0
	fi

	for dir in "$HOME/.local/bin" "$HOME/bin" /usr/local/bin; do
		if [ -f "$dir/$CMD" ]; then
			echo "$dir/$CMD"
			return 0
		fi
	done

	# Not installed. Explicit success: a failing status here would abort the
	# caller under `set -e`.
	return 0
}

# The binary path this script last installed to, or nothing. A compiled binary
# carries no comment to mark as ours, so the install records where it went.
recorded_binary() {
	src_dir=$(resolve_src_dir)
	if [ -f "$src_dir/installed-at" ]; then
		cat "$src_dir/installed-at"
	fi
}

record_install() {
	src_dir=$(resolve_src_dir)
	mkdir -p "$src_dir"
	echo "$1" >"$src_dir/installed-at"
	echo "$2" >"$src_dir/installed-ref"
}

# True when the directory is a checkout of this project rather than some
# unrelated path handed to CYBERSPACE_SRC_DIR.
is_our_source() {
	[ -f "$1/go.mod" ] || return 1
	grep -q 'cyberspacecli' "$1/go.mod" 2>/dev/null
}

# A short description of what is installed, for the install/update messages.
installed_ref() {
	src_dir=$(resolve_src_dir)
	if [ -f "$src_dir/installed-ref" ]; then
		cat "$src_dir/installed-ref"
		return 0
	fi
	echo "$REF"
}

# uname's spelling of the platform, translated to Go's. Returns non-zero for
# anything releases are not built for — that is a reason to build from source,
# not to give up, so it is not an error.
platform() {
	os=$(uname -s | tr '[:upper:]' '[:lower:]')
	case "$os" in
	linux) os=linux ;;
	darwin) os=darwin ;;
	mingw* | msys* | cygwin*) os=windows ;;
	*) return 1 ;;
	esac

	arch=$(uname -m)
	case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) return 1 ;;
	esac

	echo "${os}_${arch}"
}

# Install a published release binary. Returns non-zero when there isn't one to
# install, which is not an error: the caller falls back to a source build.
try_release() {
	install_dir="$1"

	if [ -n "${CYBERSPACE_FROM_SOURCE:-}" ]; then
		return 1
	fi

	target=$(platform) || {
		say "no release binaries are built for this platform — building from source"
		return 1
	}
	asset="${CMD}_${target}"
	case "$target" in
	windows_*) asset="${asset}.exe" ;;
	esac

	# A ref that isn't a version tag is a branch, and branches have no release
	# assets — go straight to source rather than 404 first.
	case "$REF" in
	v[0-9]*) url="https://github.com/$REPO/releases/download/$REF/$asset" ;;
	main | latest) url="https://github.com/$REPO/releases/latest/download/$asset" ;;
	*) return 1 ;;
	esac

	tmpdir=$(mktemp -d)
	if ! fetch "$url" "$tmpdir/$CMD"; then
		rm -rf "$tmpdir"
		say "no release binary for $target — building from source"
		return 1
	fi

	# A repo with no releases can serve an HTML error page with a 200, and a
	# 400-byte "binary" is a confusing way to fail. Anything this small is not a
	# Go binary.
	if [ "$(wc -c <"$tmpdir/$CMD")" -lt 500000 ]; then
		rm -rf "$tmpdir"
		say "the downloaded release asset does not look like a binary — building from source"
		return 1
	fi

	say "downloaded $asset"
	chmod 755 "$tmpdir/$CMD"
	install_binary "$tmpdir/$CMD" "$install_dir"
	rm -rf "$tmpdir"

	case "$REF" in
	main | latest) record_install "$install_dir/$CMD" "latest release" ;;
	*) record_install "$install_dir/$CMD" "$REF" ;;
	esac
}

build_from_source() {
	install_dir="$1"
	src_dir=$(resolve_src_dir)

	have go || err "Go is required to build $CMD from source.
       Install it from https://go.dev/doc/install and re-run, or wait for a
       release binary for your platform."

	fetch_source "$src_dir"

	say "building with $(go version | cut -d' ' -f3)"
	tmp_binary=$(mktemp)
	rm -f "$tmp_binary" # `go build` wants to create the file itself.
	# Quiet on success, but show Go's own output when it fails — a broken build
	# is the one time the user needs the compiler's error.
	if ! (cd "$src_dir" && go build -o "$tmp_binary" . >/dev/null 2>&1); then
		(cd "$src_dir" && go build -o "$tmp_binary" .) || err "build failed in $src_dir"
	fi

	chmod 755 "$tmp_binary"
	install_binary "$tmp_binary" "$install_dir"
	record_install "$install_dir/$CMD" "$(source_ref "$src_dir")"
}

# Put the source at $1, updating it in place when it is already there. Uses git
# when available and falls back to a tarball, so this works on minimal images
# that do not ship git.
fetch_source() {
	dir="$1"
	mkdir -p "$(dirname "$dir")"

	if [ -d "$dir" ] && [ -e "$dir/go.mod" ] && ! is_our_source "$dir"; then
		err "$dir exists and is not a cyberspacecli checkout — move it or set CYBERSPACE_SRC_DIR"
	fi

	if [ -d "$dir/.git" ] && have git; then
		say "updating source in $dir"
		(
			cd "$dir"
			git fetch --depth 1 origin "$REF" >/dev/null 2>&1 ||
				err "could not fetch $REF from origin"
			# Detached, and hard: the installed copy is not a working tree, so
			# local edits are discarded rather than left to conflict on the next
			# update.
			git checkout -q --force --detach FETCH_HEAD >/dev/null 2>&1 ||
				err "could not check out $REF"
			# Keep the install record; it lives in this directory too.
			git clean -qfd -e installed-at -e installed-ref >/dev/null 2>&1 || true
		)
		return
	fi

	if have git; then
		say "downloading source ($REPO@$REF)"
		preserve_record "$dir"
		rm -rf "$dir"
		git clone --depth 1 --branch "$REF" \
			"https://github.com/$REPO.git" "$dir" >/dev/null 2>&1 ||
			err "clone failed — check that $REPO@$REF exists"
		restore_record "$dir"
		return
	fi

	fetch_tarball "$dir"
}

fetch_tarball() {
	dir="$1"
	need_cmd tar
	say "downloading source tarball ($REPO@$REF)"

	tmpdir=$(mktemp -d)
	trap 'rm -rf "$tmpdir"' EXIT INT TERM

	fetch "https://codeload.github.com/$REPO/tar.gz/$REF" "$tmpdir/src.tar.gz" ||
		err "download failed — check that $REPO@$REF exists"

	mkdir -p "$tmpdir/src"
	tar -xzf "$tmpdir/src.tar.gz" -C "$tmpdir/src" --strip-components=1 ||
		err "could not extract the source tarball"

	is_our_source "$tmpdir/src" || err "the downloaded archive is not cyberspacecli"

	# Replace the tree wholesale: without git there is nothing to merge against,
	# and a stale file left behind is worse than a slower install.
	preserve_record "$dir"
	rm -rf "$dir"
	mkdir -p "$dir"
	# Copied with tar rather than `mv`/`cp -r` so the dotfiles come along too.
	(cd "$tmpdir/src" && tar -cf - .) | (cd "$dir" && tar -xf -) ||
		err "could not move the source into $dir"
	restore_record "$dir"

	rm -rf "$tmpdir"
	tmpdir=""
	trap - EXIT INT TERM
}

# The install record lives alongside the source, so a re-fetch that replaces the
# tree has to carry it across or `remove` would forget what it installed.
preserve_record() {
	record_backup=""
	if [ -f "$1/installed-at" ]; then
		record_backup=$(mktemp)
		cat "$1/installed-at" >"$record_backup"
	fi
}

restore_record() {
	if [ -n "${record_backup:-}" ] && [ -f "$record_backup" ]; then
		mkdir -p "$1"
		cat "$record_backup" >"$1/installed-at"
		rm -f "$record_backup"
	fi
	record_backup=""
}

# What got built, for the install message: a tag when the checkout is on one, a
# short commit otherwise.
source_ref() {
	if [ -d "$1/.git" ] && have git; then
		ref=$(cd "$1" && git describe --tags --always 2>/dev/null || true)
		if [ -n "$ref" ]; then
			echo "$ref"
			return 0
		fi
	fi
	echo "$REF"
}

install_binary() {
	binary="$1"
	install_dir="$2"

	mkdir -p "$install_dir" 2>/dev/null || true
	if [ -w "$install_dir" ]; then
		# Replacing a running binary fails on some systems with ETXTBSY; moving
		# the old one aside first does not.
		rm -f "$install_dir/$CMD" 2>/dev/null || true
		mv -f "$binary" "$install_dir/$CMD"
	elif have sudo; then
		say "$install_dir is not writable — installing with sudo"
		sudo mkdir -p "$install_dir"
		sudo rm -f "$install_dir/$CMD" 2>/dev/null || true
		sudo mv -f "$binary" "$install_dir/$CMD"
		sudo chmod 755 "$install_dir/$CMD"
	else
		rm -f "$binary"
		err "cannot write to $install_dir — set CYBERSPACE_INSTALL_DIR to a writable directory"
	fi
}

remove_path() {
	if [ -w "$(dirname "$1")" ]; then
		rm -f "$1"
	elif have sudo; then
		say "$(dirname "$1") is not writable — removing with sudo"
		sudo rm -f "$1"
	else
		err "cannot remove $1 — no write permission and sudo is unavailable"
	fi
}

# fetch URL DEST — returns non-zero on HTTP errors; callers report the reason,
# so the downloader's own stderr is suppressed.
fetch() {
	if have curl; then
		curl -fsSL "$1" -o "$2" 2>/dev/null
	elif have wget; then
		wget -q "$1" -O "$2" 2>/dev/null
	else
		err "need curl or wget to download files"
	fi
}

warn_path() {
	case ":$PATH:" in
	*":$1:"*) ;;
	*) say "note: $1 is not on your PATH — add it with: export PATH=\"$1:\$PATH\"" ;;
	esac
}

main "$@"
