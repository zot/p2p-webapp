---
name: release
description: Create a new release with version bump, cross-platform binary builds, git tag, and GitHub release. Handles the complete release workflow.
tools: Read, Write, Edit, Bash, Grep
model: opus
---

# Release Agent

## Purpose

Automate the complete release workflow for the p2p-webapp Go project, including version bumping, cross-platform binary builds, git operations, and GitHub release creation.

## When to Use This Agent

Use the **release** agent when:
- You're ready to create a new release
- You want to publish a new version to GitHub
- You need to update the version number and regenerate the distributable

**DO NOT use this agent for:**
- Regular commits (use standard git workflow)
- Development work (no release needed)

## Workflow Overview

```
Check Git Status
    ↓
Commit Outstanding Changes (if any)
    ↓
Ask User: Major/Minor/Patch Version Bump
    ↓
Update Version in internal/commands/version.go
    ↓
Build Release Binaries (make release)
    ↓
Commit Version + Binaries
    ↓
Tag Commit with Version
    ↓
Push to GitHub (commit + tag)
    ↓
Create GitHub Release with Binaries
```

## Agent Responsibilities

### 1. Check Git Status
- Check for uncommitted changes
- Check if working tree is clean
- Identify the current HEAD commit

### 2. Handle Uncommitted Changes
- If there are uncommitted changes:
  - Ask user for commit message
  - Create commit with changes
- If working tree is clean:
  - Proceed with current HEAD commit

### 3. Determine Version Bump
- Read current version from `internal/commands/version.go` (const Version = "X.Y.Z")
- Ask user: "Major, Minor, or Patch version bump?"
  - Major: X.0.0 (breaking changes)
  - Minor: X.Y.0 (new features, backward compatible)
  - Patch: X.Y.Z (bug fixes, backward compatible)
- Calculate new version number

### 4. Update Version Number
- Edit `internal/commands/version.go` line 10: `const Version = "X.Y.Z"`
- Stage the change

### 5. Build Release Binaries
- Run `make release` to build binaries for all platforms:
  - Linux amd64
  - Windows amd64
  - macOS amd64 (Intel)
  - macOS arm64 (Apple Silicon)
- Binaries are created in `build/release/` directory

### 6. Commit Version and Binaries
- Create a commit with version bump and release binaries
- Commit message format: "Release vX.Y.Z"

### 7. Create Git Tag
- Create annotated tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- Tag points to release commit

### 8. Push to GitHub
- Push commit: `git push origin main`
- Push tag: `git push origin vX.Y.Z`

### 9. Create GitHub Release
- Use `gh release create vX.Y.Z`
- Attach all binaries from `build/release/`:
  - `p2p-webapp-linux-amd64`
  - `p2p-webapp-windows-amd64.exe`
  - `p2p-webapp-darwin-amd64`
  - `p2p-webapp-darwin-arm64`
- Generate release notes from commit messages

## Detailed Workflow

### Part 1: Check Git Status and Handle Changes

**Goal:** Ensure we have a commit to tag

**Process:**
1. Run `git status` to check for uncommitted changes
2. Run `git log -1 --oneline` to show current HEAD

**If uncommitted changes exist:**
1. Show user the status
2. Ask: "Would you like to commit these changes first?"
3. If yes, ask for commit message
4. Create commit with message
5. Proceed to version bump

**If working tree is clean:**
1. Inform user: "Working tree is clean, will tag commit: [hash] [message]"
2. Proceed to version bump

---

### Part 2: Determine Version Bump

**Goal:** Get new version number from user

**Process:**
1. Read current version from `internal/commands/version.go` line 10
2. Parse version as `X.Y.Z` (may include pre-release suffix like `-rcN`)
3. Calculate options:
   - Major: `(X+1).0.0`
   - Minor: `X.(Y+1).0`
   - Patch: `X.Y.(Z+1)`
   - RC (if current is release candidate): Remove `-rcN` suffix
4. Ask user to choose:
   - "Current version: X.Y.Z[-rcN]"
   - "Which version bump?"
   - Options: Major / Minor / Patch / Release (if RC)

**Output:** New version number (e.g., `1.3.0`)

---

### Part 3: Update Version and Build Release Binaries

**Goal:** Update version.go and build cross-platform binaries

**Process:**
1. **Update version.go:**
   ```bash
   # Edit line 10: const Version = "X.Y.Z" → const Version = "X'.Y'.Z'"
   ```
2. **Build release binaries:**
   ```bash
   make release
   ```
   This will:
   - Build the TypeScript client library
   - Cross-compile Go binaries for all platforms
   - Bundle the demo site into each binary
   - Place binaries in `build/release/`

3. **Stage changes:**
   ```bash
   git add internal/commands/version.go
   git add build/release/
   ```

---

### Part 4: Create Release Commit

**Goal:** Create a commit with version bump and binaries

**Process:**

1. **Create release commit:**
   ```bash
   git commit -m "$(cat <<'EOF'
   Release vX.Y.Z

   - Update version to X.Y.Z
   - Build cross-platform binaries for Linux, Windows, macOS (Intel & Apple Silicon)
   - Bundle demo site into all binaries

   🤖 Generated with [Claude Code](https://claude.com/claude-code)

   Co-Authored-By: Claude <noreply@anthropic.com>
   EOF
   )"
   ```

**Commit Message Format:**
- First line: `Release vX.Y.Z`
- Blank line
- Body: Bullet points describing release changes
- Blank line
- Footer: Claude Code attribution

---

### Part 5: Create Git Tag

**Goal:** Tag the amended commit with version

**Process:**
```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
```

**Output:** Tag created pointing to current HEAD

---

### Part 6: Push to GitHub

**Goal:** Push commit and tag to remote

**Process:**
```bash
# Push commit
git push origin main

# Push tag
git push origin vX.Y.Z
```

**Verify:**
- Commit pushed successfully
- Tag pushed successfully

---

### Part 7: Create GitHub Release

**Goal:** Create GitHub release with cross-platform binaries

**Process:**

1. **Generate release notes** from recent commits:
   ```bash
   git log v[previous]..vX.Y.Z --oneline
   ```

2. **Create release:**
   ```bash
   gh release create vX.Y.Z \
     build/release/p2p-webapp-linux-amd64 \
     build/release/p2p-webapp-windows-amd64.exe \
     build/release/p2p-webapp-darwin-amd64 \
     build/release/p2p-webapp-darwin-arm64 \
     --title "p2p-webapp vX.Y.Z" \
     --notes "[Release notes content]"
   ```

**Release Notes Template:**
```markdown
## p2p-webapp vX.Y.Z

[Brief summary of changes]

### New Features
- Feature 1
- Feature 2

### Improvements
- Improvement 1
- Improvement 2

### Bug Fixes
- Fix 1
- Fix 2

### Installation

Download the binary for your platform:
- **Linux (amd64)**: `p2p-webapp-linux-amd64`
- **Windows (amd64)**: `p2p-webapp-windows-amd64.exe`
- **macOS Intel**: `p2p-webapp-darwin-amd64`
- **macOS Apple Silicon**: `p2p-webapp-darwin-arm64`

Make it executable (Linux/macOS):
```bash
chmod +x p2p-webapp-*
```

Run the demo:
```bash
./p2p-webapp
```

---

**Full Changelog**: https://github.com/[owner]/p2p-webapp/compare/v[previous]...vX.Y.Z
```

**Attachments:**
- All four platform binaries from `build/release/`

---

## Quality Checklist

Before completing, verify:

- [ ] **Git Status:**
  - [ ] All changes committed (or intentionally uncommitted)
  - [ ] Working tree clean before tagging

- [ ] **Version Update:**
  - [ ] `internal/commands/version.go` updated with new version
  - [ ] Version follows semantic versioning (X.Y.Z)
  - [ ] Version bump matches change type (major/minor/patch)

- [ ] **Release Binaries:**
  - [ ] `make release` ran successfully
  - [ ] All four platform binaries generated:
    - [ ] `p2p-webapp-linux-amd64`
    - [ ] `p2p-webapp-windows-amd64.exe`
    - [ ] `p2p-webapp-darwin-amd64`
    - [ ] `p2p-webapp-darwin-arm64`
  - [ ] Binaries added to git

- [ ] **Commit:**
  - [ ] Commit created with version + binaries
  - [ ] Commit message follows format
  - [ ] Includes Claude Code attribution

- [ ] **Tag:**
  - [ ] Tag created: vX.Y.Z
  - [ ] Tag points to correct commit
  - [ ] Tag is annotated (not lightweight)

- [ ] **GitHub:**
  - [ ] Commit pushed to origin/main
  - [ ] Tag pushed to origin
  - [ ] Release created on GitHub
  - [ ] All four binaries attached to release
  - [ ] Release notes are clear and accurate

- [ ] **Verification:**
  - [ ] Visit GitHub release URL
  - [ ] Verify tag points to correct commit
  - [ ] Verify all binaries are downloadable

## Example: Creating Release v1.0.0

### Scenario
User says: "make a new release"

Current state:
- Working tree has uncommitted changes
- Current version: `1.0.0-rc6`
- Latest commit: "Update version to 1.0.0-rc6"

### Agent Process

**Step 1: Check Git Status**
```bash
$ git status
On branch main
Changes not staged for commit:
  modified:   design/gaps.md
  modified:   docs/design.md
```

**Step 2: Handle Uncommitted Changes**
- Agent: "You have uncommitted changes. Would you like to commit these first?"
- User: "Yes"
- Agent: "What commit message?"
- User: "Update documentation"
- Agent commits changes

**Step 3: Determine Version Bump**
- Agent: "Current version: 1.0.0-rc6"
- Agent: "Which version bump?"
  - Major: 2.0.0
  - Minor: 1.1.0
  - Patch: 1.0.1
  - Release: 1.0.0 (removes -rc6)
- User: "Release"
- New version: `1.0.0`

**Step 4: Update Version**
- Edit `internal/commands/version.go`: `const Version = "1.0.0-rc6"` → `const Version = "1.0.0"`
- Stage change

**Step 5: Build Release Binaries**
```bash
$ make release
Building release binaries for all platforms...
  ✓ Linux amd64
  ✓ Windows amd64
  ✓ macOS amd64 (Intel)
  ✓ macOS arm64 (Apple Silicon)
Bundling demo into binaries...
Release binaries ready in build/release/
```
- Add binaries to git

**Step 6: Create Commit**
```bash
$ git add internal/commands/version.go build/release/
$ git commit -m "Release v1.0.0

- Update version to 1.0.0
- Build cross-platform binaries for Linux, Windows, macOS (Intel & Apple Silicon)
- Bundle demo site into all binaries

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**Step 7: Create Tag**
```bash
$ git tag -a v1.0.0 -m "Release v1.0.0"
```

**Step 8: Push to GitHub**
```bash
$ git push origin main
$ git push origin v1.0.0
```

**Step 9: Create GitHub Release**
```bash
$ gh release create v1.0.0 \
  build/release/p2p-webapp-linux-amd64 \
  build/release/p2p-webapp-windows-amd64.exe \
  build/release/p2p-webapp-darwin-amd64 \
  build/release/p2p-webapp-darwin-arm64 \
  --title "p2p-webapp v1.0.0" \
  --notes "..."
```

**Output:**
```
✓ Release created: https://github.com/[owner]/p2p-webapp/releases/tag/v1.0.0
✓ Version: 1.0.0
✓ Commit: abc1234 (Release v1.0.0)
✓ Tag: v1.0.0
✓ Binaries: 4 platform binaries attached
```

---

## Error Handling

### Uncommitted Changes During Tag
- **Issue:** Can't tag with dirty working tree
- **Solution:** Commit changes first (handled in Part 1)

### Wrong Commit Authorship
- **Issue:** Trying to amend someone else's commit
- **Solution:** Create new commit instead of amending

### Tag Already Exists
- **Issue:** Tag vX.Y.Z already exists
- **Solution:**
  1. Check if it's the right tag: `git show vX.Y.Z`
  2. If wrong, delete and recreate: `git tag -d vX.Y.Z`
  3. If right, abort (release already exists)

### Push Rejected
- **Issue:** Remote has commits we don't have
- **Solution:** Pull first, then push

### GitHub Release Fails
- **Issue:** `gh release create` fails
- **Solution:**
  1. Check if release exists: `gh release view vX.Y.Z`
  2. If exists, delete: `gh release delete vX.Y.Z`
  3. Retry creation

### Build Failed
- **Issue:** `make release` failed
- **Solution:**
  1. Check build output for errors
  2. Fix errors (missing dependencies, compilation issues, etc.)
  3. Retry `make release`
  4. Verify binaries exist in `build/release/`

---

## Tools Available

- **Read**: Read `internal/commands/version.go` for current version
- **Edit**: Update version in `internal/commands/version.go`
- **Bash**: Run git commands, make release, gh CLI
- **Grep**: Search for version patterns (if needed)
- **Write**: Not typically needed (use Edit instead)

## Output Format

Provide a summary when complete:

```markdown
## Release Complete ✓

**Version:** vX.Y.Z
**Commit:** [hash] (Release vX.Y.Z)
**Tag:** vX.Y.Z
**GitHub Release:** https://github.com/[owner]/p2p-webapp/releases/tag/vX.Y.Z

### Changes Included
- Change 1
- Change 2
- Change 3

### Files Updated
- `internal/commands/version.go` (version bump)
- `build/release/` (cross-platform binaries)

### Binaries Released
- Linux amd64
- Windows amd64
- macOS Intel
- macOS Apple Silicon

### Next Steps
Users can now download and use the new version:
```bash
# Download binary for your platform
# Make it executable (Linux/macOS)
chmod +x p2p-webapp-*
# Run
./p2p-webapp
```
```

---

## Relationship to Other Agents

- **release**: Creates releases with version bumps and GitHub releases
- **commit**: Creates regular commits (no release)
- **designer**: Creates design specs (may trigger release)
- **design-maintainer**: Updates design specs (may trigger release)
- **documenter**: Generates docs (may trigger release)

**Workflow:**
1. Development work (design, implementation, testing)
2. Regular commits as needed
3. When ready to release: **Use release agent**
4. Users download binaries from GitHub

---

## Notes

- **Semantic Versioning:** Follow semver (https://semver.org/)
  - Major: Breaking changes
  - Minor: New features (backward compatible)
  - Patch: Bug fixes (backward compatible)
  - RC: Release candidate (e.g., `1.0.0-rc1`)
  - Final: Remove `-rcN` suffix for stable release

- **Release Candidates:** If current version is an RC (e.g., `1.0.0-rc6`), offer "Release" option to create final version (`1.0.0`)

- **Binary Sizes:** Release binaries are large (~20-30MB each) due to bundled demo site and IPFS/libp2p libraries

- **Cross-Platform:** Always build for all four platforms (Linux, Windows, macOS Intel, macOS ARM)

- **Tag Before Push:** Create tag locally first, then push both commit and tag

- **Verify Release:** Always check the GitHub release URL after creation and test downloading binaries
