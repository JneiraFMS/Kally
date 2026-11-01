#!/usr/bin/env bash
# build-css.sh — régénère web/static/tailwind.css (classes Tailwind utilisées
# dans les sources Go + palette Solarized de tailwind.config.js) puis
# l'embarque dans le binaire via go:embed (web/assets.go).
#
# Usage : ./tools/build-css.sh [--watch]
# Dépendance : le CLI Tailwind standalone (pas besoin de Node), version
# fixée, mise en cache dans $CACHE.
set -euo pipefail

TTY_VERSION="v3.4.17"
CACHE="${HOME}/.cache/grz"
BIN="${CACHE}/tailwindcss"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

mkdir -p "${CACHE}"
if [[ ! -x "${BIN}" ]]; then
	echo "Téléchargement de tailwindcss ${TTY_VERSION} (CLI standalone)…"
	case "$(uname -m)" in
		x86_64)  ASSET="tailwindcss-linux-x64" ;;
		aarch64) ASSET="tailwindcss-linux-arm64" ;;
		*) echo "arch non supportée"; exit 1 ;;
	esac
	curl -fsSL -o "${BIN}" \
		"https://github.com/tailwindlabs/tailwindcss/releases/download/${TTY_VERSION}/${ASSET}"
	chmod +x "${BIN}"
fi

cd "${REPO_ROOT}"

if [[ "${1:-}" == "--watch" ]]; then
	exec "${BIN}" -c tailwind.config.js -i web/static/input.css \
		-o web/static/tailwind.css --watch --minify
fi

"${BIN}" -c tailwind.config.js -i web/static/input.css \
	-o web/static/tailwind.css --minify

echo "Généré : web/static/tailwind.css ($(wc -c < web/static/tailwind.css) octets)"
