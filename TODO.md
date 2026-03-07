# Dotfiles Project TODO

## 🚀 High Priority: GitHub Automation
- [ ] **Release Workflow**: Implement `.github/workflows/release.yml` for cross-platform Go builds.
- [ ] **`install.sh` Integration**: Update the legacy script to auto-detect OS/Arch and download pre-compiled TUI binaries.
- [ ] **Versioning**: Implement semver tagging and SHA256 checksum generation for releases.

## 🛠️ TUI Configuration Screens
- [ ] **SSH Config Screen**: Automate SSH key generation and `~/.ssh/config` management.
- [ ] **Shell Overrides Screen**: Add a UI for managing machine-specific aliases in `~/.zshrc.local`.
- [ ] **GPG Signing Screen**: Streamline the setup of Git commit signing.

## 📝 Documentation & Housekeeping
- [x] **Centralized TODO**: Create `TODO.md` (this file).
- [ ] **Plan Archive**: Move completed plans in `docs/plans/` to a `completed/` subdirectory.
- [ ] **README Update**: Finalize `README.md` to highlight the new TUI installer as the primary setup method.

## 🎨 TUI Polish & UX
- [x] **Centralized Status Bar**: Centralized footer rendering in `app.go`.
- [x] **Error Handling**: Improved visual reporting for build/stow failures in Summary screen.
- [ ] **Input Validation**: Add more robust regex validation for email and 1Password references in config screens.
