#!/usr/bin/env bash
# One entry point for a fresh installation or restoration of a full fork backup.
set -Eeuo pipefail

usage() {
    cat <<'EOF'
Fresh installation:
  sudo bash setup.sh [release-tag]
Restore a trusted full backup on Ubuntu/Debian:
  sudo bash setup.sh --backup /path/full-backup.tar.gz [--check] [--replace]
--check validates the archive and compatibility without changing services/files.
--replace explicitly permits replacing an existing installation.
Use the same distribution version and CPU architecture as the backup server.
EOF
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    usage
    exit 0
fi
[[ $EUID -eq 0 ]] || { echo 'Run with sudo or as root.' >&2; exit 1; }

backup=''
tag='dev-latest'
restore_args=()
while (($#)); do
    case "$1" in
        --backup)
            [[ $# -ge 2 && -f "$2" ]] || { echo '--backup requires a local archive.' >&2; exit 1; }
            backup="$2"; shift 2 ;;
        --check|--replace) restore_args+=("$1"); shift ;;
        -*) usage >&2; exit 1 ;;
        *) [[ $# -eq 1 && -z "$backup" ]] || { usage >&2; exit 1; }
           tag="$1"; shift ;;
    esac
done
[[ -n "$backup" || ${#restore_args[@]} -eq 0 ]] || { usage >&2; exit 1; }

work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
chmod 700 "$work"
base='https://raw.githubusercontent.com/martynov-dev/3x-ui/main'

if [[ -n "$backup" ]]; then
    # No installer runs here: the exact panel executable is already in the backup.
    if ! command -v python3 >/dev/null; then
        command -v apt-get >/dev/null || { echo 'Ubuntu/Debian with apt is required.' >&2; exit 1; }
        apt-get update
        apt-get install -y python3
    fi
    local_helper="$(dirname -- "${BASH_SOURCE[0]}")/scripts/restore_backup.py"
    if [[ -f "$local_helper" ]]; then
        helper="$local_helper"
    else
        command -v curl >/dev/null || { apt-get update; apt-get install -y curl ca-certificates; }
        curl -fsSL --retry 3 "$base/scripts/restore_backup.py" -o "$work/restore_backup.py"
        helper="$work/restore_backup.py"
    fi
    python3 "$helper" --backup "$backup" "${restore_args[@]}"
else
    [[ -r /etc/os-release ]] || { echo 'Ubuntu/Debian is required.' >&2; exit 1; }
    # shellcheck disable=SC1091
    source /etc/os-release
    [[ "${ID:-}" == ubuntu || "${ID:-}" == debian ]] || { echo 'Ubuntu/Debian is required.' >&2; exit 1; }
    apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y curl ca-certificates nginx openssh-client
    curl -fsSL --retry 3 "$base/install.sh" -o "$work/install.sh"
    XUI_NONINTERACTIVE=1 bash "$work/install.sh" "$tag"
fi
