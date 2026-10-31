#!/usr/bin/env bash
# FixLayout Go installer for Ubuntu / Linux Mint on X11/Xorg
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
BIN="$HOME/.local/bin/fixlayout"
KEY="${KEY:-<Primary><Alt>k}"
CHAR_KEY="${CHAR_KEY:-<Primary>BackSpace}"
DESKTOP="${XDG_CURRENT_DESKTOP:-}"
if [[ "${1:-}" == "--uninstall" ]]; then
  rm -f "$BIN"
  echo "Removed $BIN. Remove shortcuts from desktop settings if needed."
  exit 0
fi
if [[ $# -gt 0 && "${1:-}" != "--no-deps" ]]; then
  echo "Usage: ./install.sh [--no-deps|--uninstall]" >&2
  exit 2
fi
if [[ "${1:-}" != "--no-deps" ]]; then
  sudo apt-get update -qq
  sudo apt-get install -y golang-go build-essential libxkbcommon0 xkb-data xclip xdotool libnotify-bin hunspell-en-us
  sudo apt-get install -y hunspell-th 2>/dev/null || echo "Optional hunspell-th unavailable; using Thai fallback"
fi
"$HERE/build.sh"
mkdir -p "$HOME/.local/bin"
if [[ -f "$BIN" && ! -f "$BIN.before-go.bak" ]]; then
  cp -p "$BIN" "$BIN.before-go.bak"
fi
install -m 755 "$HERE/bin/fixlayout-go" "$BIN"
if [[ "${XDG_SESSION_TYPE:-}" == "wayland" ]]; then echo "Warning: X11/Xorg session required"; fi
add_binding() {
  local id="$1" name="$2" cmd="$3" bind="$4" cur new P K SCHEMA
  if [[ "$DESKTOP" == *Cinnamon* ]]; then
    SCHEMA="org.cinnamon.desktop.keybindings"
    P="/org/cinnamon/desktop/keybindings/custom-keybindings/$id/"
    cur=$(gsettings get "$SCHEMA" custom-list)
    if [[ "$cur" != *"'$id'"* ]]; then
      if [[ "$cur" == '@as []' || "$cur" == '[]' ]]; then new="['$id']"; else new="${cur%]}, '$id']"; fi
      gsettings set "$SCHEMA" custom-list "$new"
    fi
    K="$SCHEMA.custom-keybinding:$P"
    gsettings set "$K" binding "['$bind']"
  else
    SCHEMA="org.gnome.settings-daemon.plugins.media-keys"
    P="/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/$id/"
    cur=$(gsettings get "$SCHEMA" custom-keybindings)
    if [[ "$cur" != *"/$id/"* ]]; then
      if [[ "$cur" == '@as []' || "$cur" == '[]' ]]; then new="['$P']"; else new="${cur%]}, '$P']"; fi
      gsettings set "$SCHEMA" custom-keybindings "$new"
    fi
    K="$SCHEMA.custom-keybinding:$P"
    gsettings set "$K" binding "$bind"
  fi
  gsettings set "$K" name "$name"
  gsettings set "$K" command "$cmd"
}
if command -v gsettings >/dev/null 2>&1; then
  add_binding fixlayout fixlayout "$BIN" "$KEY"
  if [[ "$CHAR_KEY" != "none" ]]; then
    add_binding fixlayout-char fixlayout-char "$BIN --char" "$CHAR_KEY"
  fi
else
  echo "gsettings not available; configure keyboard shortcuts manually"
fi
echo "Installed to $BIN"
echo "Selection/line: $KEY | Previous character: $CHAR_KEY"
