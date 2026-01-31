# go-update Reference Documentation

This directory contains complete reference documentation for the go-update library.

## Files

### go-update-README.md
The official go-update library README containing:
- Overview and features
- Basic usage example
- API compatibility promises
- Breaking changes history
- License information

**When to read:** For general overview and feature list.

### doc.go
Complete package documentation with detailed examples:
- Basic HTTP update example
- Binary patching example
- Checksum verification example
- Cryptographic signature verification example
- Single-file binary requirements
- Non-goals and limitations

**When to read:** When implementing specific features like patching, checksums, or signatures.

## Quick Navigation

**Need to implement basic updates?** → See SKILL.md "Basic Update Pattern" section

**Need checksum verification?** → See doc.go "Checksum Verification" section

**Need signature verification?** → See doc.go "Cryptographic Signature Verification" section

**Need binary patches?** → See doc.go "Binary Patching" section

**Need complete example?** → See SKILL.md "Complete Self-Update Implementation" section
