---
description: Enforces maintaining release notes for every change in the homelab repository
---

# Release Notes Documentation Skill

## Purpose

This skill ensures that all changes to the homelab repository are properly documented in release notes **before every git commit**. This creates a comprehensive audit trail and knowledge base for future reference.

## When to Use

**MANDATORY**: This skill must be applied **before every `git commit`** command in the homelab repository.

## Release Notes Location

```
/Volumes/work/git-repos/homelab/docs/releases/v{VERSION}-RELEASE-NOTES.md
```

Current version: `v0.1.5`

## Workflow

### 1. Before Making Any Code Changes

Check if release notes file exists:
```bash
ls -la /Volumes/work/git-repos/homelab/docs/releases/
```

If the file doesn't exist for the current version, create it using the template below.

### 2. After Making Code Changes (Before Commit)

**REQUIRED STEPS**:

1. **Update the "Changes Log" section** with:
   - Date and time (use current timestamp)
   - Files modified
   - Problem description
   - Solution implemented
   - Debugging steps taken (if applicable)

2. **Update the "Commits" section** with:
   - Commit hash (after commit)
   - Commit message
   - Files changed
   - Brief description

3. **Update relevant sections**:
   - Applications Modified (if Helm charts updated)
   - Infrastructure Changes (if infrastructure modified)
   - Breaking Changes (if applicable)
   - Known Issues (if new issues discovered)

### 3. Commit Pattern

```bash
# 1. Make code changes
# 2. Update release notes
git add docs/releases/v{VERSION}-RELEASE-NOTES.md
git add <other-files>

# 3. Commit with descriptive message
git commit -m "type: description

- Detail 1
- Detail 2

Updates release notes"

# 4. Push
git push origin <branch>
```

## Release Notes Template

```markdown
# Release v{VERSION} - {Title}

**Release Date**: {Date}  
**Branch**: `v{VERSION}`  
**Status**: In Progress | Complete

## Overview

Brief description of this release's focus.

## Changes Log

### {Date} {Time} - {Component/Feature}

**Problem**:
- Description of the issue

**Solution**:
- What was changed
- Why it was changed

**Files Modified**:
- `path/to/file1.go`
- `path/to/file2.go`

**Debugging Steps** (if applicable):
1. Step 1
2. Step 2

**Commit**: `{hash}` - {message}

---

## Applications Modified

| Application | Change Type | Details |
|-------------|-------------|---------|
| App Name | Version/Config | Description |

## Infrastructure Changes

### {Component Name}

**Changes**:
- Change 1
- Change 2

**Configuration**:
```yaml
# Relevant config
```

## Commits

- `{hash}` - {message}
  - Files: `file1.go`, `file2.go`
  - Description: Brief description

## Known Issues

### {Issue Title}

**Description**: Issue description

**Impact**: Impact description

**Workaround**: Workaround if available

## Verification Steps

```bash
# Commands to verify changes
```

## References

- [Link 1](url)
- [Link 2](url)
```

## Example Entry

```markdown
### 2026-02-08 11:00 - Longhorn Node Selector Fix

**Problem**:
- Longhorn manager DaemonSet showing 0/0 pods
- Worker nodes don't have `node-role.kubernetes.io/worker` label
- DaemonSet couldn't match any nodes

**Solution**:
- Removed incorrect node selectors from longhornManager, longhornDriver, longhornUI
- Added tolerations to avoid control-plane nodes instead
- Allows Longhorn to schedule on all worker nodes

**Files Modified**:
- `platform/cdk8s/cots/storage/longhorn.go`

**Debugging Steps**:
1. Checked DaemonSet status: `kubectl get daemonset -n longhorn-system`
2. Verified node labels: `kubectl get nodes --show-labels`
3. Found workers have no `node-role.kubernetes.io/worker` label
4. Replaced node selectors with tolerations

**Commit**: `609bcd5` - fix: Remove incorrect worker node selector from Longhorn
```

## Enforcement Rules

### ❌ NEVER commit without updating release notes

**Bad**:
```bash
git add file.go
git commit -m "fix something"
git push
```

**Good**:
```bash
# 1. Make changes to file.go
# 2. Update release notes
git add docs/releases/v0.1.5-RELEASE-NOTES.md
git add file.go
git commit -m "fix: description

Updates release notes"
git push
```

### ✅ Always include

1. **Date and time** of the change
2. **Problem description** - what was broken
3. **Solution description** - what was fixed
4. **Files modified** - complete list
5. **Debugging steps** - if troubleshooting was involved
6. **Commit reference** - link to the commit

### ✅ Update these sections as needed

- **Applications Modified** - for Helm chart updates
- **Infrastructure Changes** - for infrastructure modifications
- **Breaking Changes** - for breaking changes
- **Known Issues** - for newly discovered issues
- **Verification Steps** - for testing procedures

## Integration with Git Workflow

### Pre-Commit Checklist

Before running `git commit`:

- [ ] Release notes file exists for current version
- [ ] Added new entry to "Changes Log" section
- [ ] Updated "Commits" section (can be done after commit)
- [ ] Updated relevant sections (Applications, Infrastructure, etc.)
- [ ] Staged release notes file: `git add docs/releases/v{VERSION}-RELEASE-NOTES.md`

### Post-Commit Update

After `git commit`:

1. Get commit hash: `git log -1 --format=%h`
2. Update "Commits" section with hash
3. Amend commit if needed: `git commit --amend --no-edit`

## Benefits

1. **Complete Audit Trail**: Every change is documented with context
2. **Knowledge Transfer**: Future developers understand why changes were made
3. **Debugging Reference**: Troubleshooting steps are preserved
4. **Release Management**: Easy to generate release notes
5. **Long-term Memory**: AI assistants can reference past decisions

## Version Management

### When to Create New Release Notes

Create a new release notes file when:
- Starting a new version (e.g., v0.1.6)
- Major feature branch
- Significant refactoring effort

### File Naming Convention

```
v{MAJOR}.{MINOR}.{PATCH}-RELEASE-NOTES.md
```

Examples:
- `v0.1.5-RELEASE-NOTES.md`
- `v0.2.0-RELEASE-NOTES.md`
- `v1.0.0-RELEASE-NOTES.md`

## Automation Opportunities

Future enhancements:
- Pre-commit hook to check if release notes updated
- Script to generate commit entries automatically
- Template generator for new releases
- Changelog generator from release notes

---

**Remember**: Release notes are not optional. They are a critical part of the development workflow and must be maintained for every change.
