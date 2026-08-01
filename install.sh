#!/usr/bin/env sh
#
# install.sh - install the cyberspace TUI without cloning the repository.
#
#   ./install.sh           install the latest published release into ~/.local/bin
#   ./install.sh v1.2.3    install a specific release tag
#   ./install.sh --source  build and install from source (requires Go)
#
set -eu

VERSION=""
SOURCE=0
for arg in "$@"; do
	case "$arg" in
	--source)
		SOURCE=1
		;;
	-h | --help)
		sed -n '2,5p' "$0"
		exit 0
		;;
	*)
		VERSION="$arg"
		;;
	esac
done

REPO="johannesalke/cyberspacecli"
BIN="cyberspace"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
x86_64 | amd64) arch="amd64" ;;
aarch64 | arm64) arch="arm64" ;;
esac

if [ "$SOURCE" = "1" ]; then
	echo "Installing from source..."
	exec go install "github.com/$REPO/cmd/cyberspace@latest"
fi

if [ -z "$VERSION" ]; then
	VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)"
fi

if [ -z "$VERSION" ]; then
	echo "Could not resolve a release; falling back to source install."
	exec go install "github.com/$REPO/cmd/cyberspace@latest"
fi

url="https://github.com/$REPO/releases/download/$VERSION/$BIN-$os-$arch.tar.gz"
echo "Downloading $url"
mkdir -p "$HOME/.local/bin"
curl -fsSL "$url" | tar -xz -C "$HOME/.local/bin" "$BIN"
chmod +x "$HOME/.local/bin/$BIN"
echo "Installed $BIN to $HOME/.local/bin/$BIN"
echo "Add $HOME/.local/bin to your PATH if it is not already there."
