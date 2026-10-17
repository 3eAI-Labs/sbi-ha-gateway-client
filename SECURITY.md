# Security Policy

## Supported versions

Only the most recent release receives fixes.

## Reporting a vulnerability

**Do not open a public issue.**

Use GitHub's private vulnerability reporting:
[**Report a vulnerability**](https://github.com/3eAI-Labs/sbi-ha-gateway-client/security/advisories/new).
The report stays private between you and the maintainers until a fix is published.

Include the module version, what an attacker gains, and a reproduction — ideally a failing test.
If the problem is in the gateway rather than in this module, report it to
[sbi-ha-gateway](https://github.com/3eAI-Labs/sbi-ha-gateway/security/advisories/new) instead.

## Scope

In scope: a schema that makes a failure indistinguishable from a success, a field whose documented
encoding differs from what the gateway reads or writes, and anything in this module that would
cause key material (authentication vectors) or subscriber identities to be exposed.

Out of scope: the NATS server, its permissions, and the transport security of your connection —
see the Security section of the package documentation for how to restrict the subjects.
