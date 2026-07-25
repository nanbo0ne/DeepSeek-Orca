# Security Policy

## Supported Releases

Security fixes are provided for the latest stable DeepSeek-Orca desktop
release. Users should update to the newest release before reporting an issue
that may already have been corrected.

## Reporting A Vulnerability

Please use GitHub's private vulnerability-reporting form:

<https://github.com/nanbo0ne/DeepSeek-Orca/security/advisories/new>

Include the affected version, operating system, reproduction steps, expected
impact, and any relevant logs with credentials and personal data removed. Do
not publish exploit details, signing abuse, leaked credentials, or other
sensitive evidence in a public issue.

If private reporting is unavailable, open a minimal GitHub issue asking a
maintainer for a private contact channel. Do not include vulnerability details
in that issue.

Maintainers will acknowledge a complete report, assess its impact, and
coordinate a fix and disclosure timeline with the reporter. Reports involving
official Windows binaries or Authenticode signatures are also handled under
the controls documented in [SIGNING.md](SIGNING.md).

## Release Integrity

Official downloads are published from this repository's GitHub Releases page.
Windows release artifacts must carry a valid, timestamped Authenticode
signature. Do not run an artifact whose signature is missing or invalid.
