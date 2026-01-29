# Agent Container Utilities

This document provides a comprehensive reference of utilities available in the agent container and their library dependencies. It serves as a cross-reference companion to [Agent Container System Libraries](./agent-container-system-libraries.md).

## Overview

The agent container includes a variety of utilities for development workflows, categorized by function. This document maps each utility to its required system libraries from `/lib`, `/lib64`, and `/usr/lib` directories.

## Utility Categories

### Shell Environments

#### bash
- **Path:** `/usr/bin/bash`
- **Purpose:** GNU Bourne-Again Shell - interactive command interpreter
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libtinfo.so.6` - Terminal capability database
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** None
- **Library Category:** Critical only
- **Use Cases:**
  - Interactive shell sessions
  - Shell script execution
  - Command-line workflows
- **Testing:** `bash --version`, `bash -c 'echo "test"'`

#### sh
- **Path:** `/usr/bin/sh` (typically dash or bash symlink)
- **Purpose:** POSIX shell - minimal shell for scripts
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** None
- **Library Category:** Critical only
- **Use Cases:**
  - POSIX-compliant scripts
  - Minimal shell environments
  - System initialization scripts
- **Testing:** `sh --version`, `sh -c 'echo "test"'`

### Version Control

#### git
- **Path:** `/usr/bin/git`
- **Purpose:** Distributed version control system
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libpcre2-8.so.0` - Pattern matching for regex operations
  - `libz.so.1` - Object compression/decompression
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** (from `/usr/lib`)
  - `libgit2.so` - Git core library (if available)
- **Library Category:** Critical only
- **Use Cases:**
  - Clone repositories
  - Commit changes
  - Push/pull operations
  - Branch management
  - History inspection
- **Testing:** 
  ```bash
  git --version
  git init /tmp/test-repo
  git -C /tmp/test-repo status
  ```

### Networking & HTTP

#### curl
- **Path:** `/usr/bin/curl`
- **Purpose:** Transfer data with URLs - HTTP, HTTPS, FTP, and more
- **System Libraries:** (from `/lib` and `/lib64`)
  - **Essential:**
    - `libcurl.so.4` - HTTP client library
    - `libz.so.1` - Compression support
    - `libc.so.6` - C standard library
    - `ld-linux-x86-64.so.2` - Dynamic linker
  - **Security (HTTPS):**
    - `libssl.so.3` - SSL/TLS protocol
    - `libcrypto.so.3` - Cryptographic functions
  - **Protocols:**
    - `libnghttp2.so.14` - HTTP/2 support
    - `libssh.so.4` - SSH/SFTP support
    - `librtmp.so.1` - RTMP streaming (optional)
  - **Authentication:**
    - `libgssapi_krb5.so.2` - Kerberos/GSSAPI
    - `libkrb5.so.3` - Kerberos 5
    - `libk5crypto.so.3` - Kerberos crypto
    - `libkrb5support.so.0` - Kerberos support
    - `libldap.so.2` - LDAP protocol (optional)
    - `liblber.so.2` - LDAP BER encoding (optional)
    - `libsasl2.so.2` - SASL mechanisms (optional)
  - **Compression:**
    - `libzstd.so.1` - Zstandard compression (optional)
    - `libbrotlidec.so.1` - Brotli decompression (optional)
    - `libbrotlicommon.so.1` - Brotli common (optional)
  - **Internationalization:**
    - `libidn2.so.0` - Internationalized domain names (optional)
    - `libunistring.so.5` - Unicode string handling (optional)
  - **Other:**
    - `libpsl.so.5` - Public suffix list for cookie validation
    - `libresolv.so.2` - DNS resolution
    - `libcom_err.so.2` - Error description
    - `libkeyutils.so.1` - Kernel key management (optional)
  - **TLS Alternative:**
    - `libgnutls.so.30` - GnuTLS library (optional)
    - `libhogweed.so.6` - Crypto algorithms (optional)
    - `libnettle.so.8` - Crypto primitives (optional)
    - `libtasn1.so.6` - ASN.1 parsing (optional)
    - `libp11-kit.so.0` - PKCS#11 support (optional)
    - `libffi.so.8` - Foreign function interface (optional)
    - `libgmp.so.10` - Multi-precision arithmetic (optional)
- **User Libraries:** None (libcurl provides all functionality)
- **Library Category:** Critical + Important + Optional (for full feature set)
- **Use Cases:**
  - HTTP API calls
  - File downloads
  - HTTPS communication
  - OAuth/authentication workflows
  - Multi-protocol transfers
- **Testing:**
  ```bash
  curl --version
  curl -I https://github.com
  curl -v https://api.github.com 2>&1 | grep "TLS"
  ```

**Note:** curl has the most extensive library dependencies of any utility, supporting multiple protocols, authentication methods, and compression algorithms.

### Data Processing

#### jq
- **Path:** `/usr/bin/jq`
- **Purpose:** Command-line JSON processor
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libjq.so.1` - JSON parsing and query execution
  - `libm.so.6` - Math functions for numeric operations
  - `libonig.so.5` - Regular expression support
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** None
- **Library Category:** Critical + Important
- **Use Cases:**
  - Parse JSON responses
  - Transform JSON data
  - Extract values from JSON
  - Filter and map JSON arrays
- **Testing:**
  ```bash
  jq --version
  echo '{"name":"test","value":42}' | jq '.name'
  echo '[1,2,3]' | jq 'map(. * 2)'
  ```

#### yq
- **Path:** `/usr/bin/yq` (if installed)
- **Purpose:** YAML processor (similar to jq for YAML)
- **System Libraries:** Varies by implementation (Go binary or Python wrapper)
  - Go version: Minimal dependencies (statically linked or standard Go runtime)
  - Python version: Same as python3 + PyYAML dependencies
- **Use Cases:**
  - Parse YAML configuration files
  - Transform YAML data
  - Extract values from YAML
- **Testing:** `yq --version`, `echo 'key: value' | yq '.key'`

### Text Processing

#### grep
- **Path:** `/usr/bin/grep`
- **Purpose:** Search text using patterns
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libpcre2-8.so.0` - Perl-compatible regular expressions
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** None
- **Library Category:** Critical only
- **Use Cases:**
  - Search files for patterns
  - Filter command output
  - Regular expression matching
- **Testing:**
  ```bash
  grep --version
  echo -e "foo\nbar\nbaz" | grep "ba"
  ```

#### awk
- **Path:** `/usr/bin/awk` (typically gawk)
- **Purpose:** Pattern scanning and text processing
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libsigsegv.so.2` - Segmentation fault recovery (optional)
  - `libreadline.so.8` - Interactive line editing
  - `libmpfr.so.6` - Multiple-precision floating-point
  - `libgmp.so.10` - Multiple-precision arithmetic
  - `libm.so.6` - Math functions
  - `libc.so.6` - C standard library
  - `libtinfo.so.6` - Terminal info
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** None
- **Library Category:** Critical + Important + Optional
- **Use Cases:**
  - Text column extraction
  - Pattern-based processing
  - Arithmetic operations
  - Report generation
- **Testing:**
  ```bash
  awk --version
  echo -e "1 2 3\n4 5 6" | awk '{print $1 + $2}'
  ```

#### sed
- **Path:** `/usr/bin/sed`
- **Purpose:** Stream editor for text transformation
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libacl.so.1` - POSIX ACL support
  - `libselinux.so.1` - SELinux security contexts
  - `libpcre2-8.so.0` - Regular expression support
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** None
- **Library Category:** Critical + Important
- **Use Cases:**
  - Text substitution
  - Line filtering
  - In-place file editing
- **Testing:**
  ```bash
  sed --version
  echo "hello world" | sed 's/world/universe/'
  ```

### File Operations

#### tar
- **Path:** `/usr/bin/tar`
- **Purpose:** Archive utility for creating and extracting tar files
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libacl.so.1` - Preserve POSIX ACLs
  - `libselinux.so.1` - Preserve SELinux contexts
  - `libpcre2-8.so.0` - Pattern matching
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** None
- **Library Category:** Critical + Important
- **Use Cases:**
  - Create archives
  - Extract archives
  - List archive contents
  - Compress/decompress with gzip/bzip2
- **Testing:**
  ```bash
  tar --version
  tar -czf /tmp/test.tar.gz /etc/hostname
  tar -tzf /tmp/test.tar.gz
  ```

#### gzip
- **Path:** `/usr/bin/gzip`
- **Purpose:** Compress or expand files
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** None
- **Library Category:** Critical only
- **Use Cases:**
  - Compress files
  - Decompress files
  - Stream compression
- **Testing:**
  ```bash
  gzip --version
  echo "test data" | gzip | gunzip
  ```

### Programming Runtimes

#### python3
- **Path:** `/usr/bin/python3`
- **Purpose:** Python interpreter
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libm.so.6` - Math operations
  - `libz.so.1` - Compression (zlib module)
  - `libexpat.so.1` - XML parsing
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** (from `/usr/lib`)
  - Python standard library modules (many .so files)
  - Additional C extensions as needed
- **Library Category:** Critical + Important
- **Use Cases:**
  - Run Python scripts
  - Python-based tools
  - Data processing
  - Automation scripts
- **Testing:**
  ```bash
  python3 --version
  python3 -c "import sys; print(sys.version)"
  python3 -c "import json; print(json.dumps({'test': 1}))"
  ```

#### node
- **Path:** `/opt/hostedtoolcache/node/*/x64/bin/node` or `/usr/bin/node`
- **Purpose:** Node.js JavaScript runtime
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libdl.so.2` - Dynamic loading for native modules
  - `libstdc++.so.6` - C++ standard library
  - `libm.so.6` - Math operations
  - `libgcc_s.so.1` - GCC runtime support
  - `libpthread.so.0` - Threading support
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** (from `/usr/lib`)
  - Node.js built-in modules
  - Native addons as needed
- **Library Category:** Critical only
- **Use Cases:**
  - Run JavaScript applications
  - Execute npm scripts
  - Node.js-based tools
  - Build processes
- **Testing:**
  ```bash
  node --version
  node -e "console.log('Hello from Node.js')"
  node -e "console.log(process.version)"
  ```

#### go
- **Path:** `/opt/hostedtoolcache/go/*/x64/bin/go` or `/usr/bin/go`
- **Purpose:** Go compiler and tool
- **System Libraries:** (from `/lib` and `/lib64`)
  - **None** - Go binaries are typically statically linked
- **User Libraries:** None
- **Library Category:** N/A (statically linked)
- **Use Cases:**
  - Compile Go programs
  - Run Go tools
  - Execute go commands
- **Testing:**
  ```bash
  go version
  go env GOVERSION
  ```

**Note:** Go binaries are statically linked by default, making them independent of system libraries (except for certain cgo-based operations).

#### java
- **Path:** `/usr/bin/java`
- **Purpose:** Java Virtual Machine
- **System Libraries:** (from `/lib` and `/lib64`)
  - `libpthread.so.0` - Threading support
  - `libdl.so.2` - Dynamic loading for JNI
  - `libc.so.6` - C standard library
  - `ld-linux-x86-64.so.2` - Dynamic linker
- **User Libraries:** (from `/usr/lib`)
  - Java runtime libraries (rt.jar, etc.)
  - JNI libraries
- **Library Category:** Critical only
- **Use Cases:**
  - Run Java applications
  - Execute JAR files
  - Java-based tools
- **Testing:**
  ```bash
  java --version
  java -version 2>&1 | head -1
  ```

### GitHub Integration

#### gh
- **Path:** `/usr/bin/gh`
- **Purpose:** GitHub CLI - GitHub command-line tool
- **System Libraries:** Depends on implementation
  - Go binary: Minimal (statically linked or Go runtime only)
  - May depend on git for some operations
- **User Libraries:** None (statically linked Go binary)
- **Library Category:** Minimal
- **Use Cases:**
  - Repository operations
  - Issue management
  - Pull request workflows
  - GitHub API access
- **Testing:**
  ```bash
  gh --version
  gh auth status
  ```

## Utility-to-Library Matrix

This table shows which utilities require which library categories:

| Utility | Critical | Important | Optional | Notes |
|---------|----------|-----------|----------|-------|
| bash | ✓ | - | - | Shell + basic C library |
| sh | ✓ | - | - | Minimal shell |
| git | ✓ | - | - | VCS operations |
| curl | ✓ | ✓ | ✓ | Full networking stack |
| jq | ✓ | ✓ | - | JSON + math |
| grep | ✓ | - | - | Pattern matching |
| awk | ✓ | ✓ | ✓ | Text processing + math |
| sed | ✓ | ✓ | - | Stream editing + ACL |
| tar | ✓ | ✓ | - | Archiving + ACL |
| gzip | ✓ | - | - | Compression only |
| python3 | ✓ | ✓ | - | Runtime + modules |
| node | ✓ | - | - | JavaScript runtime |
| go | - | - | - | Statically linked |
| java | ✓ | - | - | JVM + JNI |
| gh | - | - | - | Statically linked |

## Library Requirements by Workflow

### Basic Shell Workflow
**Utilities:** bash, sh, grep, sed, gzip  
**Required Libraries:** Critical only (10 libraries)
```
/lib64/ld-linux-x86-64.so.2
libc.so.6
libm.so.6
libdl.so.2
libpthread.so.0
libz.so.1
libpcre2-8.so.0
libtinfo.so.6
libgcc_s.so.1
libstdc++.so.6
```

### Git Workflow
**Utilities:** bash, git, grep, sed, tar, gzip  
**Required Libraries:** Critical only (10 libraries) - same as basic shell
**Use Cases:**
- Clone repositories
- Commit and push changes
- Branch operations
- Local version control

### API/Network Workflow
**Utilities:** bash, curl, jq  
**Required Libraries:** Critical + Important (25 libraries)
**Additional Libraries:**
- All crypto/SSL libraries
- All Kerberos libraries
- Network protocol libraries
**Use Cases:**
- GitHub API calls
- REST API integration
- HTTPS downloads
- Authentication workflows

### Data Processing Workflow
**Utilities:** bash, jq, python3, grep, awk, sed  
**Required Libraries:** Critical + Important (25 libraries)
**Use Cases:**
- JSON data manipulation
- CSV processing
- Log analysis
- Report generation

### Development Workflow
**Utilities:** bash, git, node, python3, jq, curl  
**Required Libraries:** Critical + Important (25 libraries)
**Use Cases:**
- Full development environment
- Build processes
- Testing
- CI/CD pipelines

## Cross-Reference with System Libraries

For detailed information about each library's purpose, size, and security considerations, see [Agent Container System Libraries](./agent-container-system-libraries.md).

### Quick Reference

**Critical Libraries (10):**
- Always required for basic container operation
- See: [Critical Libraries](./agent-container-system-libraries.md#critical-libraries-container-wont-function)

**Important Libraries (15):**
- Required for networking, security, and common utilities
- See: [Important Libraries](./agent-container-system-libraries.md#important-libraries-common-utilities-need)

**Optional Libraries (21):**
- Specialized protocols and functionality
- See: [Optional Libraries](./agent-container-system-libraries.md#optional-libraries-specialized-use-cases)

## Mounting Recommendations by Use Case

### Minimal Shell Environment
```yaml
volumes:
  # Mount only critical libraries
  - /lib/x86_64-linux-gnu/libc.so.6:/lib/x86_64-linux-gnu/libc.so.6:ro
  - /lib/x86_64-linux-gnu/libm.so.6:/lib/x86_64-linux-gnu/libm.so.6:ro
  - /lib/x86_64-linux-gnu/libdl.so.2:/lib/x86_64-linux-gnu/libdl.so.2:ro
  - /lib/x86_64-linux-gnu/libpthread.so.0:/lib/x86_64-linux-gnu/libpthread.so.0:ro
  - /lib/x86_64-linux-gnu/libz.so.1:/lib/x86_64-linux-gnu/libz.so.1:ro
  - /lib/x86_64-linux-gnu/libpcre2-8.so.0:/lib/x86_64-linux-gnu/libpcre2-8.so.0:ro
  - /lib/x86_64-linux-gnu/libtinfo.so.6:/lib/x86_64-linux-gnu/libtinfo.so.6:ro
  - /lib/x86_64-linux-gnu/libgcc_s.so.1:/lib/x86_64-linux-gnu/libgcc_s.so.1:ro
  - /lib/x86_64-linux-gnu/libstdc++.so.6:/lib/x86_64-linux-gnu/libstdc++.so.6:ro
  - /lib64/ld-linux-x86-64.so.2:/lib64/ld-linux-x86-64.so.2:ro
```

**Supported Utilities:** bash, sh, grep, gzip, basic git

### Standard Development Environment (Recommended)
```yaml
volumes:
  # Mount entire directories for simplicity
  - /lib/x86_64-linux-gnu:/lib/x86_64-linux-gnu:ro
  - /lib64:/lib64:ro
```

**Supported Utilities:** All utilities with full functionality

**Benefits:**
- Simple configuration
- Handles all dependencies automatically
- Supports symlinks
- Future-proof

**Drawbacks:**
- Larger attack surface (1.2 GB)
- Includes unused libraries

### Production Hardened Environment
```yaml
volumes:
  # Mount only what's needed based on utility requirements
  # Start with critical, add important as needed
  - /lib/x86_64-linux-gnu:/lib/x86_64-linux-gnu:ro
  - /lib64:/lib64:ro
  
# Additional security
securityContext:
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

**Recommendation:** Full directory mounting with strict security context

## Troubleshooting

### Missing Library Errors

**Error Pattern:**
```
error while loading shared libraries: libXXX.so.Y: cannot open shared object file
```

**Resolution Steps:**
1. Identify the utility throwing the error
2. Use `ldd /path/to/utility` to see all dependencies
3. Check which libraries are from `/lib` or `/lib64`
4. Add missing libraries to mount configuration
5. Verify symlinks are preserved

**Example:**
```bash
# Find missing library
ldd /usr/bin/curl | grep "not found"

# Check where library should be
find /lib /lib64 -name "libssl.so*" 2>/dev/null

# Add to mount config
- /lib/x86_64-linux-gnu/libssl.so.3:/lib/x86_64-linux-gnu/libssl.so.3:ro
```

### Version Mismatch

**Symptom:** Utility works on host but fails in container

**Common Causes:**
- Container libraries are older/newer than expected
- Symlink pointing to wrong version
- ABI compatibility issues

**Resolution:**
- Mount entire directory to preserve version symlinks
- Use container-native library installation
- Update host libraries to match container expectations

### Performance Issues

**Symptom:** Utilities run slower than expected

**Possible Causes:**
- Network mount latency (if using remote volumes)
- Many small library files
- Excessive symlink resolution

**Resolution:**
- Copy libraries into container image for performance
- Use bind mounts instead of volume mounts
- Profile with `strace` to identify bottlenecks

## Testing Utilities

Comprehensive test script to validate utility functionality:

```bash
#!/bin/bash
# test-utilities.sh - Validate agent container utilities

echo "## Utility Functionality Test"
echo ""

# Shell environments
echo "### Shell Environments"
bash --version | head -1 && echo "✅ bash" || echo "❌ bash"
sh --version 2>&1 | head -1 && echo "✅ sh" || echo "❌ sh"
echo ""

# Version control
echo "### Version Control"
git --version && echo "✅ git" || echo "❌ git"
echo ""

# Networking
echo "### Networking"
curl --version | head -1 && echo "✅ curl" || echo "❌ curl"
curl -Is https://github.com | head -1 && echo "✅ curl HTTPS" || echo "❌ curl HTTPS"
echo ""

# Data processing
echo "### Data Processing"
jq --version && echo "✅ jq" || echo "❌ jq"
echo '{"test":1}' | jq . > /dev/null && echo "✅ jq processing" || echo "❌ jq processing"
echo ""

# Text processing
echo "### Text Processing"
grep --version | head -1 && echo "✅ grep" || echo "❌ grep"
awk --version | head -1 && echo "✅ awk" || echo "❌ awk"
sed --version | head -1 && echo "✅ sed" || echo "❌ sed"
echo ""

# File operations
echo "### File Operations"
tar --version | head -1 && echo "✅ tar" || echo "❌ tar"
gzip --version | head -1 && echo "✅ gzip" || echo "❌ gzip"
echo ""

# Runtimes
echo "### Programming Runtimes"
python3 --version && echo "✅ python3" || echo "❌ python3"
node --version && echo "✅ node" || echo "❌ node"
go version && echo "✅ go" || echo "❌ go"
java --version 2>&1 | head -1 && echo "✅ java" || echo "❌ java"
echo ""

# GitHub integration
echo "### GitHub Integration"
gh --version && echo "✅ gh" || echo "❌ gh"
echo ""
```

Save and run:
```bash
chmod +x test-utilities.sh
./test-utilities.sh
```

## Related Documentation

- [Agent Container System Libraries](./agent-container-system-libraries.md) - Detailed library analysis
- [Agent Container Testing](./agent-container-testing.md) - Smoke test workflow
- Issue #11971: Utility audit baseline
- Issue #11972: `/usr/lib` shared libraries analysis
- Issue #11970: Overall container hardening

---

**Document Version:** 1.0  
**Last Updated:** 2026-01-29  
**Author:** GitHub Copilot Agent  
**Related Issues:** #11970, #11971, #11972
