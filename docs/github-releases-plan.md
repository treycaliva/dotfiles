# GitHub Releases Plan for Go TUI Installer

## Goal

Automatically build cross-compiled Go binaries and publish them as GitHub Release assets when a version is tagged.

## Target Platforms

| GOOS/GOARCH     | Covers                        |
|-----------------|-------------------------------|
| `darwin/arm64`  | macOS Apple Silicon           |
| `linux/amd64`   | Linux x86_64 + WSL2           |
| `linux/arm64`   | Linux ARM (Raspberry Pi, etc) |

## Workflow: `.github/workflows/release.yml`

### Trigger

Push of a semver tag:

```yaml
on:
  push:
    tags:
      - 'v*'
```

### Steps

1. **Checkout** the repo
2. **Set up Go** (match version from `go.mod`, currently 1.24.2)
3. **Build matrix** — cross-compile for each platform:
   ```bash
   GOOS=darwin  GOARCH=arm64 go build -o installer-darwin-arm64  ./cmd/installer
   GOOS=linux   GOARCH=amd64 go build -o installer-linux-amd64   ./cmd/installer
   GOOS=linux   GOARCH=arm64 go build -o installer-linux-arm64   ./cmd/installer
   ```
4. **Create GitHub Release** using `softprops/action-gh-release` (or `gh release create`)
5. **Attach binaries** as release assets

### Binary naming convention

```
installer-{os}-{arch}
```

### Download URLs (after release)

Latest release:
```
https://github.com/treycaliva/dotfiles/releases/latest/download/installer-darwin-arm64
https://github.com/treycaliva/dotfiles/releases/latest/download/installer-linux-amd64
https://github.com/treycaliva/dotfiles/releases/latest/download/installer-linux-arm64
```

## install.sh Integration

Add a function to `install.sh` that:

1. Detects OS (`uname -s`) and arch (`uname -m`)
2. Maps to the correct binary name:
   - `Darwin` + `arm64` -> `installer-darwin-arm64`
   - `Linux` + `x86_64` -> `installer-linux-amd64`
   - `Linux` + `aarch64` -> `installer-linux-arm64`
3. Downloads from the latest release URL using `curl -fsSL`
4. Makes it executable and runs it

Example snippet:
```sh
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64)  arch="amd64" ;;
  aarch64) arch="arm64" ;;
esac

url="https://github.com/treycaliva/dotfiles/releases/latest/download/installer-${os}-${arch}"
curl -fsSL -o /tmp/dotfiles-installer "$url"
chmod +x /tmp/dotfiles-installer
/tmp/dotfiles-installer
```

## How to Release

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow builds and publishes automatically. The release page and download URLs are live within a few minutes.

## Open Questions

- Should the workflow also trigger on PR merge to main (as a pre-release / `latest` tag)?
- Should checksums (SHA256) be generated and attached for verification?
- Should `install.sh` fall back to the shell-based installer if the download fails?
