# Security Policy

## Supported Versions

We follow an **End-of-Life (EOL)** policy to provide security and bug fix support for KubeSphere versions.

We regularly release patch versions to address security vulnerabilities and critical bugs for supported KubeSphere
releases. The support period for each version is determined by its **EOL date**, rather than by a fixed number of minor
versions.

The current support plan is as follows:

| KubeSphere Version            | End of Life (EOL) Date |
|-------------------------------|------------------------|
| **KubeSphere v4.2**           | ---                    |
| **KubeSphere v4.1**           | Sep 12, 2027           |
| **KubeSphere v3.4**           | Dec 25, 2025           |
| **KubeSphere v3.3 & earlier** | Oct 31, 2025           |

Once a version reaches its EOL date, it will no longer receive official security updates or bug fixes. Older versions
may receive **critical security fixes on a best-effort basis**, but we cannot guarantee that all security patches will
be backported to unsupported versions.

In rare cases, where a security fix requires significant architectural changes or is otherwise highly intrusive, and a
feasible workaround exists, we may choose to **apply the fix only in a future release**, rather than backporting it to a
patch version for currently supported releases.

For long-term stability, we recommend users plan their upgrades according to the EOL schedule.

Let me know if you'd like any refinements!

## Reporting a Vulnerability

# Security Vulnerability Disclosure and Response Process

To ensure KubeSphere security, a security vulnerability disclosure and response process is adopted. And the security team is set up in KubeSphere community, also any issue and PR is welcome for every contributors.

The primary goal of this process is to reduce the total exposure time of users to publicly known vulnerabilities. To quickly fix vulnerabilities of KubeSphere, the security team is responsible for the entire vulnerability management process, including internal communication and external disclosure.

If you find a vulnerability or encounter a security incident involving vulnerabilities of KubeSphere, please report it as soon as possible to the KubeSphere security team (security@kubesphere.io).

Please kindly help provide as much vulnerability information as possible in the following format:

- Issue title (Please add `Security` label)*:

- Overview*:

- Affected components and version number*:

- CVE number (if any):

- Vulnerability verification process*:

- Contact information*:

The asterisk (*) indicates the required field.

# Response Time

The KubeSphere security team will confirm the vulnerabilities and contact you within 2 working days after your submission.

We will publicly thank you after fixing the security vulnerability. To avoid negative impact, please keep the vulnerability confidential until we fix it. We would appreciate it if you could obey the following code of conduct:

The vulnerability will not be disclosed until KubeSphere releases a patch for it.

The details of the vulnerability, for example, exploits code, will not be disclosed.
