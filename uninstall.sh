#!/usr/bin/env bash
set -euo pipefail

removed=false
if [ -f /usr/local/bin/keptler ]; then
	rm -f /usr/local/bin/keptler
	echo "Removed /usr/local/bin/keptler"
	removed=true
fi
if [ -f "$HOME/.local/bin/keptler" ]; then
	rm -f "$HOME/.local/bin/keptler"
	echo "Removed $HOME/.local/bin/keptler"
	removed=true
fi
if [ "$removed" = false ]; then
	echo "keptler binary not found in standard locations" >&2
fi
