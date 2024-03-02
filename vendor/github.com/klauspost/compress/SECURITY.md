# Security Policy

## Supported Versions

Security updates are applied only to the latest release.

## Vulnerability Definition

A security vulnerability is a bug that with certain input triggers a crash or an infinite loop. Most calls will have varying execution time and only in rare cases will slow operation be considered a security vulnerability.

Corrupted output generally is not considered a security vulnerability, unless independent operations are able to affect each other. Note that not all functionality is re-entrant and safe to use concurrently.

Out-of-memory crashes only applies if the en/decoder uses an abnormal amount of memory, with appropriate options applied, to limit maximum window size, concurrency, etc. However, if you are in doubt you are welcome to file a security issue.

It is assumed that all callers are trusted, meaning internal data exposed through reflection or inspection of returned data structures is not considered a vulnerability.

Vulnerabilities resulting from compiler/assembler errors should be reported upstream. Depending on the severity this package may or may not implement a workaround.

## Reporting a Vulnerability

If you have discovered a security vulnerability in this project, please report it privately. **Do not disclose it as a public issue.** This gives us time to work with you to fix the issue before public exposure, reducing the chance that the exploit will be used before a patch is released.

<<<<<<< HEAD
<<<<<<< HEAD
Please disclose it at [security advisory](https://github.com/klauspost/compress/security/advisories/new). If possible please provide a minimal reproducer. If the issue only applies to a single platform, it would be helpful to provide access to that.
=======
Please disclose it at [security advisory](https://github.com/klaupost/compress/security/advisories/new). If possible please provide a minimal reproducer. If the issue only applies to a single platform, it would be helpful to provide access to that.
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
Please disclose it at [security advisory](https://github.com/klauspost/compress/security/advisories/new). If possible please provide a minimal reproducer. If the issue only applies to a single platform, it would be helpful to provide access to that.
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

This project is maintained by a team of volunteers on a reasonable-effort basis. As such, vulnerabilities will be disclosed in a best effort base.
