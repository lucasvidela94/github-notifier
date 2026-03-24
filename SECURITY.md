# Security Policy

## Supported Versions

The following versions of github-notifier receive security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please send an email to
lucasvidela94@gmail.com. Do not open a public issue.

Please include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fixes (optional)

We aim to respond within 48 hours with an initial assessment, and provide
a timeline for the fix.

## Security Best Practices

- GitHub tokens are stored in `~/.config/github-notifier/env` with `600` permissions
- The token is only used for GitHub API calls
- No data is sent to third-party servers
- All communication is directly with GitHub's API

## Scope

This security policy applies to the github-notifier application and its
official distribution methods (GitHub releases, AUR package).
