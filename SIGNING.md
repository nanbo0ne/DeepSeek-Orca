# Windows Code-Signing Policy

O.R.C.A's official Windows application and NSIS installer are built from
the tagged source in this public repository and signed through SignPath.

Free code signing is provided by [SignPath.io](https://signpath.io/), with a
publicly trusted certificate supplied by the
[SignPath Foundation](https://signpath.org/). The certificate private key stays
in SignPath's hardware-backed signing service and is never available to project
maintainers or GitHub Actions runners.

## Signed Artifacts

- `Orca.exe`, distributed inside the Windows portable ZIP and
  installed by the NSIS package.
- `O.R.C.A-for-Windows-windows-amd64-installer.exe`, distributed from the official
  GitHub Release.

The release workflow signs the application first, packages that signed binary,
and then signs the completed installer. Normal tag-triggered Windows releases
fail closed if SignPath is unavailable or either Authenticode signature cannot
be validated.

During initial SignPath Foundation enrollment, a maintainer may use the manual
`allow_unsigned_windows` workflow input for a temporary release. The input is
off by default, cannot be enabled by pushing a tag, and the corresponding
Release must state that its Windows files are unsigned. Once signing is
available, the same release assets are rebuilt and replaced with verified,
timestamped files.

## Trusted Build

1. A maintainer creates a `desktop-v*` tag from the default branch.
2. GitHub Actions builds the application on a GitHub-hosted Windows runner.
3. The unsigned application artifact is submitted through SignPath's GitHub
   trusted build-system connector.
4. The signed application is packaged into the NSIS installer.
5. The completed installer is submitted for a second signing request.
6. The workflow verifies signer certificates and trusted timestamps before any
   Windows artifact can reach the GitHub Release.

The SignPath project uses the committed `windows-executable` artifact
configuration from
`.signpath/artifact-configurations/windows-executable.xml`. Signing requests
must use the repository's configured release-signing policy.

## Roles And Approval

Project maintainers are responsible for reviewing release changes and creating
release tags. SignPath submitter and approver access is limited to maintainers
with multi-factor authentication. SignPath's origin verification ensures that
only artifacts produced by this repository's GitHub-hosted workflow can be
submitted under the release policy.

## Verification

On Windows, a downloaded release can be checked with:

```powershell
Get-AuthenticodeSignature .\O.R.C.A-for-Windows-windows-amd64-installer.exe |
  Format-List Status,StatusMessage,SignerCertificate,TimeStamperCertificate
```

`Status` must be `Valid`. Downloads should come only from
<https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases>.

## Reporting Problems

Do not run an artifact whose signature is missing or invalid. Report suspected
tampering or certificate misuse through the private process in
[SECURITY.md](SECURITY.md). Defender and SmartScreen false positives can also be
submitted to Microsoft's security intelligence file-submission portal.
