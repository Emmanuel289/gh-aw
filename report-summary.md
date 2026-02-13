# Firewall Escape Testing - Latest Summary

**Last Updated**: 2026-02-13T09:23:00Z  
**Run ID**: 21981347249  
**Status**: SANDBOX SECURE

## Quick Stats
- **Total Runs**: 6
- **Total Techniques Tested**: 137
- **Successful Escapes**: 1 (Run 21052141750 - Docker sibling container)
- **Latest Run**: 30 techniques, 100% novel, all blocked

## Latest Run Highlights (2026-02-13)

### Novel Approaches Tested
1. **Kernel-level**: Netlink sockets, BPF filesystem, AF_PACKET
2. **TLS/Crypto**: ALPN negotiation, certificate manipulation
3. **Squid Advanced**: Vary header cache poisoning, URL overflow, auth bypass
4. **Container Runtime**: containerd/runc socket search, cgroup manipulation
5. **Protocol Edge Cases**: HTTP/0.9, RTSP, WebSocket proper handshake
6. **Gateway Exploitation**: Port scanning, proxy attempts
7. **DNS Advanced**: CNAME chains, port 853, subdomain exfiltration
8. **Encoding Tricks**: IPv6 zone ID, Punycode, chunked extensions

### Key Findings
- Netlink sockets work (AF_NETLINK) but don't provide network escape
- Gateway has HTTP server on port 80 (returns 404, not exploitable)
- DNS queries succeed (by design, side channel only)
- All 30 techniques blocked from reaching forbidden domains

### Architecture Analysis
**Security Layers**:
1. Proxy environment variables (bypassable, ineffective alone)
2. **Squid ACL domain filtering** (primary security boundary)
3. Container network isolation
4. Limited capabilities (CAP_NET_RAW blocked)

**Traffic Flow**: Container → iptables NAT → Squid (172.30.0.10:3128) → Allowed domains only

## Recommendations for Future Runs

### Avoid Re-testing (Already Exhausted)
- Direct HTTP/HTTPS protocol bypasses
- DNS-based attacks (tunneling, TXT records, over-TCP)
- Alternative protocols (QUIC, SSH on 443, FTP, RTSP, git://, rsync)
- Application-level HTTP clients (Python, Node, Ruby, Go, Perl, PHP, Rust)
- Squid-specific attacks (cache manager, ESI, ACL bypass, CONNECT)
- Container isolation techniques (LD_PRELOAD, /proc/1/root, namespaces)
- Basic encoding tricks (URL encoding, chunked, IDN, IPv6)

### Unexplored Areas (Potential Future Focus)
1. **Time-of-check-time-of-use (TOCTOU)** - Race conditions in filter application
2. **Memory corruption** - Buffer overflows in Squid, iptables
3. **Side-channel timing** - Measure filter behavior differences
4. **Containerized tools abuse** - Other containers on awf-net (if any)
5. **Zero-day exploits** - Squid/kernel vulnerabilities (out of scope for testing)

## Historical Context

### Run Timeline
1. **20802044428** (2026-01-08): 20 basic techniques - all blocked
2. **20978685291** (2026-01-14): 20 advanced techniques - all blocked
3. **21016029008** (2026-01-15): 27 container escape attempts - all blocked
4. **21052141750** (2026-01-16): **ESCAPE FOUND** - Docker exec to node container
5. **21940571452** (2026-02-12): 30 timing/kernel techniques - all blocked
6. **21981347249** (2026-02-13): 30 kernel/protocol techniques - all blocked

### Only Successful Escape
- **Run 21052141750**: Docker exec to node:lts-alpine sibling container
- Container name: quizzical_hugle
- **Finding**: Node container had NO firewall restrictions
- **Impact**: Could access any external domain (example.com returned 200)
- **Scope**: Sibling container vulnerability, not current container

## Conclusion

The AWF firewall container (5ddc9648c878) remains secure after 6 comprehensive testing runs with 137 unique escape techniques. The Squid proxy ACL filtering is robust and effectively blocks all attempts to reach forbidden domains.

The only known escape vector (Run 21052141750) was through an unfiltered Docker sibling container, which represents a different security boundary than the current container's network isolation.
