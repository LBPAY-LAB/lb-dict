# Git Tag Instructions for v0.1.0

## Overview

This document provides instructions for creating the git tag for version v0.1.0 of the dict-contracts module.

## Status

- ✅ VERSION file created: `v0.1.0`
- ✅ CHANGELOG.md created with v0.1.0 entry
- ✅ RELEASE_NOTES.md created with detailed documentation
- ✅ README.md updated with version information
- ⏳ Git tag needs to be created manually (requires git permissions)

## Manual Git Tag Creation

### Option 1: Create Tag in dict-contracts subdirectory (Recommended)

If you want dict-contracts to be its own git repository:

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/dict-contracts

# Initialize git repository (if not already done)
git init

# Add all files
git add -A

# Create initial commit
git commit -m "Release: Version v0.1.0 - Initial release of DICT contracts

This is the initial release of the dict-contracts module, providing comprehensive
Protocol Buffers definitions for the LBPay DICT system.

Included:
- CoreDictService: 15 gRPC methods (FrontEnd <-> Core DICT)
- BridgeService: 14 gRPC methods (Connect <-> Bridge <-> Bacen)
- Common types and enums shared across all services
- Complete documentation and usage examples

Total: 29 gRPC methods across 3 proto files

Files:
- Proto files: common.proto, core_dict.proto, bridge.proto
- Documentation: VERSION, CHANGELOG.md, RELEASE_NOTES.md, README.md
- Generated code: gen/ directory
- Build config: Makefile, buf.yaml, go.mod

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Create annotated tag
git tag -a v0.1.0 -m "Initial release - DICT contracts with CoreDictService and BridgeService

Version v0.1.0 includes:
- CoreDictService: 15 gRPC methods for FrontEnd <-> Core DICT communication
- BridgeService: 14 gRPC methods for Connect <-> Bridge <-> Bacen communication
- Common types and comprehensive error handling
- Full documentation and usage examples
- 29 total gRPC methods across 3 proto files"

# Verify tag was created
git tag -l
git show v0.1.0
```

### Option 2: Create Tag in Parent Repository

If dict-contracts should remain part of the parent IA_Dict repository:

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict

# Add dict-contracts files
git add dict-contracts/

# Create commit
git commit -m "Release: dict-contracts v0.1.0 - Initial release

Initial release of dict-contracts module with comprehensive Protocol Buffers
definitions for the LBPay DICT system.

Included:
- CoreDictService: 15 gRPC methods (FrontEnd <-> Core DICT)
- BridgeService: 14 gRPC methods (Connect <-> Bridge <-> Bacen)
- Common types and enums shared across all services
- Complete documentation and usage examples

Total: 29 gRPC methods across 3 proto files

Files added:
- dict-contracts/VERSION: v0.1.0
- dict-contracts/CHANGELOG.md: Complete version history
- dict-contracts/RELEASE_NOTES.md: Detailed release documentation
- dict-contracts/README.md: Updated with version information

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Create annotated tag with module prefix
git tag -a dict-contracts/v0.1.0 -m "dict-contracts v0.1.0: Initial release

Version v0.1.0 includes:
- CoreDictService: 15 gRPC methods for FrontEnd <-> Core DICT communication
- BridgeService: 14 gRPC methods for Connect <-> Bridge <-> Bacen communication
- Common types and comprehensive error handling
- Full documentation and usage examples
- 29 total gRPC methods across 3 proto files"

# Verify tag was created
git tag -l | grep dict-contracts
git show dict-contracts/v0.1.0
```

## Verify Tag Creation

After creating the tag, verify it:

```bash
# List all tags
git tag -l

# Show tag details
git show v0.1.0

# Or if using parent repo with prefix
git show dict-contracts/v0.1.0
```

## Push Tag to Remote (Optional - for later deployment)

**Note**: DO NOT push tags to remote yet. This is for LOCAL development only.

When ready to publish (future step):

```bash
# Push specific tag
git push origin v0.1.0

# Or if using parent repo with prefix
git push origin dict-contracts/v0.1.0

# Push all tags (be careful!)
git push origin --tags
```

## Tag Information

**Tag Name**: `v0.1.0` (or `dict-contracts/v0.1.0` if in parent repo)

**Tag Message**:
```
Initial release - DICT contracts with CoreDictService and BridgeService

Version v0.1.0 includes:
- CoreDictService: 15 gRPC methods for FrontEnd <-> Core DICT communication
- BridgeService: 14 gRPC methods for Connect <-> Bridge <-> Bacen communication
- Common types and comprehensive error handling
- Full documentation and usage examples
- 29 total gRPC methods across 3 proto files
```

## Go Module Usage

Once tagged, the module can be used as:

```bash
# If dict-contracts has its own repository
go get github.com/lbpay/dict-contracts@v0.1.0

# If part of parent repository
go get github.com/lbpay/IA_Dict/dict-contracts@dict-contracts/v0.1.0
```

## Files Included in v0.1.0

```
dict-contracts/
├── VERSION                 # v0.1.0
├── CHANGELOG.md            # Version history
├── RELEASE_NOTES.md        # Detailed release documentation
├── README.md               # Updated with version info
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies
├── Makefile                # Build automation
├── buf.yaml                # Buf configuration
├── proto/
│   ├── common.proto        # Shared types
│   ├── core_dict.proto     # CoreDictService (15 methods)
│   └── bridge.proto        # BridgeService (14 methods)
├── gen/                    # Generated Go code
│   └── proto/
│       ├── common/v1/
│       ├── core/v1/
│       └── bridge/v1/
└── .github/                # CI/CD workflows
```

## Next Steps

1. ✅ Create git tag using one of the options above
2. Verify tag exists: `git tag -l`
3. Test module import in other services
4. When ready for production: Push tag to remote repository
5. Publish to internal package registry (if applicable)

## Notes

- This is a LOCAL tag only - not pushed to remote
- Use semantic versioning (v0.1.0 = initial release)
- Tag is annotated (not lightweight) for full metadata
- Tag can be deleted and recreated if needed: `git tag -d v0.1.0`
