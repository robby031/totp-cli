#!/bin/sh

set -eu

REPO="${TOTP_REPO:-robby031/totp-cli}"

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1; }

in_path() {
	case ":${PATH:-}:" in
		*":$1:"*) return 0 ;;
		*) return 1 ;;
	esac
}

pick_install_dir() {
	for dir in /usr/local/bin /opt/homebrew/bin "$HOME/.local/bin" "$HOME/bin"; do
		if [ -d "$dir" ] && [ -w "$dir" ] && in_path "$dir"; then
			printf '%s' "$dir"
			return
		fi
	done
	for dir in /usr/local/bin /opt/homebrew/bin "$HOME/.local/bin" "$HOME/bin"; do
		if [ -d "$dir" ] && [ -w "$dir" ]; then
			printf '%s' "$dir"
			return
		fi
	done
	printf '%s' "$HOME/.local/bin"
}

download() {
	url="$1"
	dest="$2"
	if need curl; then
		curl -fLsS --retry 3 -o "$dest" "$url"
	elif need wget; then
		wget -q -O "$dest" "$url"
	else
		die "butuh curl atau wget untuk mengunduh"
	fi
}

case "$(uname -s)" in
	Darwin) os="darwin" ;;
	Linux) os="linux" ;;
	*) die "OS $(uname -s) tidak didukung. Pengguna Windows: pakai install.ps1" ;;
esac

case "$(uname -m)" in
	x86_64 | amd64) arch="amd64" ;;
	arm64 | aarch64) arch="arm64" ;;
	*) die "arsitektur $(uname -m) tidak didukung" ;;
esac

if [ -n "${TOTP_VERSION:-}" ]; then
	base_url="https://github.com/$REPO/releases/download/$TOTP_VERSION"
else
	base_url="https://github.com/$REPO/releases/latest/download"
fi
base_url="${TOTP_BASE_URL:-$base_url}"

archive="totp-$os-$arch.tar.gz"

if [ -n "${TOTP_INSTALL_DIR:-}" ]; then
	install_dir="$TOTP_INSTALL_DIR"
else
	install_dir="$(pick_install_dir)"
fi

tmp="$(mktemp -d "${TMPDIR:-/tmp}/totp-cli.XXXXXX")" || die "tidak bisa membuat direktori sementara"
trap 'rm -rf "$tmp"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

say "Mengunduh totp-cli $os/$arch..."
download "$base_url/$archive" "$tmp/$archive"
download "$base_url/checksums.txt" "$tmp/checksums.txt"

expected="$(awk -v name="$archive" '$2 == name { print $1; exit }' "$tmp/checksums.txt")"
[ -n "$expected" ] || die "checksum untuk $archive tidak ditemukan"

if need sha256sum; then
	actual="$(sha256sum "$tmp/$archive" | awk '{ print $1 }')"
elif need shasum; then
	actual="$(shasum -a 256 "$tmp/$archive" | awk '{ print $1 }')"
else
	die "butuh sha256sum atau shasum untuk memverifikasi unduhan"
fi

[ "$actual" = "$expected" ] || die "checksum tidak cocok (unduhan rusak atau dimanipulasi)"

tar -xzf "$tmp/$archive" -C "$tmp"
[ -f "$tmp/totp" ] || die "binary totp tidak ditemukan di dalam arsip"

mkdir -p "$install_dir"
if need install; then
	install -m 0755 "$tmp/totp" "$install_dir/totp"
else
	cp "$tmp/totp" "$install_dir/totp"
	chmod 0755 "$install_dir/totp"
fi

version="$("$install_dir/totp" --version 2>/dev/null || true)"
say "Terpasang: $install_dir/totp ${version:+($version)}"

case ":${PATH:-}:" in
	*":$install_dir:"*) ;;
	*)
		if [ "${TOTP_NO_MODIFY_PATH:-0}" = "1" ]; then
			say "Tambahkan ke PATH: export PATH=\"$install_dir:\$PATH\""
		else
			rc=""
			shell_name="$(basename "${SHELL:-sh}")"
			case "$shell_name" in
				zsh) rc="$HOME/.zshrc" ;;
				bash)
					if [ -f "$HOME/.bash_profile" ]; then
						rc="$HOME/.bash_profile"
					else
						rc="$HOME/.bashrc"
					fi
					;;
				*) rc="$HOME/.profile" ;;
			esac

			marker="# >>> totp-cli >>>"
			if ! grep -q "$marker" "$rc" 2>/dev/null; then
				{
					printf '\n%s\n' "$marker"
					printf 'export PATH="%s:$PATH"\n' "$install_dir"
					printf '# <<< totp-cli <<<\n'
				} >>"$rc"
				say "PATH ditambahkan ke $rc"
			fi
			say "Buka terminal baru atau jalankan: export PATH=\"$install_dir:\$PATH\""
		fi
		;;
esac

say ""
if in_path "$install_dir"; then
	say "Siap digunakan. Coba: totp --version"
else
	say "Siap digunakan."
	say "Terminal baru: totp --version"
	say "Terminal ini: $install_dir/totp --version"
fi
