# Firewall Escape Testing - Latest Summary

**Last Updated**: 2026-02-14T09:15:00Z  
**Run ID**: 22014770879  
**Status**: SANDBOX SECURE

## Quick Stats
- **Total Runs**: 27
- **Total Techniques Tested**: 688
- **Successful Escapes**: 1 (Run 21052141750 - Docker sibling container, patched in v0.9.1)
- **Latest Run**: 30 techniques, 100% novel, all blocked

## Latest Run Highlights (2026-02-14)

### Novel Approaches Tested
1. **LD_PRELOAD Exploitation**: Reverse engineered one-shot-token.so library (getenv/secure_getenv hooks)
2. **host.docker.internal**: Port scanning found HTTP on port 80 (172.17.0.1), no proxy functionality
3. **Exotic Syscalls**: splice() succeeded in transferring data but still intercepted by Squid
4. **NO_PROXY Exploitation**: Confirmed NO_PROXY doesn't bypass iptables NAT (kernel level)
5. **Process Injection**: ptrace() blocked - no CAP_SYS_PTRACE
6. **Shared Memory**: /dev/shm is read-only, Unix sockets work but no network bypass
7. **Docker Embedded DNS**: Direct queries to 127.0.0.11 refused
8. **Timing Oracles**: DNS and Squid cache timing differences observable (side channel only)
9. **HTTP/0.9**: Protocol downgrade still intercepted by Squid
10. **Raw Sockets**: CAP_NET_RAW dropped, UDP to 1.1.1.1 blocked

### Key Findings
- **Kernel-level NAT enforcement**: Even exotic syscalls (splice, sendfile) are intercepted
- **All capabilities dropped**: CapEff: 0000000000000000
- **Squid 6.13 robust**: Domain filtering enforced regardless of access method
- **DNS restricted**: Only 8.8.8.8 and 8.8.4.4 allowed (1.1.1.1 times out)
- **Docker socket accessible**: But container spawn vulnerability patched in AWF v0.9.1

### Architecture Analysis
**Security Layers**:
1. **iptables NAT** (kernel level) - redirects all TCP to Squid transparently
2. **Squid 6.13 ACL** (application level) - domain whitelist filtering
3. **Capability drops** - no CAP_NET_RAW, CAP_SYS_PTRACE, CAP_SYS_ADMIN
4. **Resource restrictions** - /dev/shm read-only, chroot enabled, no-new-privs=1

**Traffic Flow**: Container → iptables NAT → Squid (172.30.0.10:3128) → Allowed domains only

## Recommendations for Future Runs

### Avoid Re-testing (Already Exhausted)
- All basic HTTP/HTTPS protocol variants and clients
- DNS-based attacks (tunneling, rebinding, over-TCP, TXT records, EDNS, CNAME chains)
- Squid-specific attacks (cache poisoning, ESI, auth bypass, CONNECT, URL overflow)
- Container isolation techniques (Docker socket spawn, namespaces, /proc/1/root, capabilities)
- Encoding tricks (URL encoding, chunked, Punycode, Unicode, IPv6 zone IDs)
- NO_PROXY and proxy environment variable manipulation (confirmed ineffective)
- Alternative protocol attempts (FTP, SSH, WebSocket, HTTP/2, QUIC, RTSP, git://, rsync)
- Alternative application clients (Python, Node, Ruby, Go, Perl, PHP, Rust, Java)
- Exotic syscalls (splice, sendfile) - both intercepted at kernel NAT level
- LD_PRELOAD exploitation - operates above NAT layer

### Unexplored Areas (Potential Future Focus)
1. **Squid zero-day exploits** - Test known CVEs in Squid 6.13
2. **iptables/netfilter kernel vulnerabilities** - Linux kernel network stack bugs
3. **Container runtime exploits** - containerd/runc specific vulnerabilities
4. **Time-of-check-time-of-use (TOCTOU)** - Advanced race conditions during startup
5. **Memory corruption** - Buffer overflows in Squid or kernel

## Historical Context

### Run Timeline (Recent)
1. **21981347249** (2026-02-13): 30 kernel/protocol techniques - all blocked
2. **21940571452** (2026-02-12): 30 timing/kernel techniques - all blocked
3. **21899322029** (2026-02-11): 30 WebDAV/DNS/kernel techniques - all blocked
4. **21858986051** (2026-02-10): 30 IPv6/SCTP/protocol techniques - all blocked
5. **22014770879** (2026-02-14): 30 LD_PRELOAD/syscall/NO_PROXY techniques - all blocked

### Only Successful Escape (Historical)
- **Run 21052141750** (2026-01-16): Docker exec to node:lts-alpine sibling container
- **Vulnerability**: Sibling container had NO firewall restrictions
- **Status**: Patched in AWF v0.9.1+
- **Scope**: Different security boundary (inter-container isolation vs network firewall)

## Conclusion

The AWF firewall container remains secure after 27 comprehensive testing runs with 688 unique escape techniques. The combination of:

1. **Kernel-level iptables NAT** (transparent redirection)
2. **Squid 6.13 ACL filtering** (domain whitelist)
3. **Strict capability drops** (CapEff: 0000000000000000)
4. **Resource restrictions** (/dev/shm read-only, chroot, no-new-privs)

...creates a robust defense-in-depth model that has successfully blocked all 688 escape attempts (excluding the one patched vulnerability in inter-container isolation).

The only known escape vector (Run 21052141750) was through an unfiltered Docker sibling container, which represents a different security boundary than the current container's network firewall.

**Firewall Status**: SECURE
