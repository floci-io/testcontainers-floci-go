# Contributing to testcontainers-floci-go

Thank you for your interest in contributing! This is a community-driven Testcontainers module for Floci and all contributions are welcome. By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

**Join us on [Slack](https://join.slack.com/t/floci/shared_invite/zt-3tjn02s3q-A00kEjJ1cZxsg_imTfy6Cw)**: it is the fastest way to reach maintainers.

## Ways to Contribute

- **Bug reports** — open an issue with a minimal reproduction
- **Feature requests** — open an issue describing the Floci service or option you need
- **Pull requests** — bug fixes, new service configurations, or improvements
- **Examples** — add usage examples under `./examples/`

## Getting Started

### Prerequisites

- Go 1.25+
- Docker (required to run the container in tests)

### Build & Test

```bash
git clone https://github.com/floci-io/testcontainers-floci-go.git
cd testcontainers-floci-go
go build ./...          # compile all packages
go test ./...           # run all tests (requires Docker)
go test -run TestRun_DefaultConfig .   # run a single test
```

## Branching Model

This module follows a **tag-driven release model**.

| Branch | Purpose | Published? |
|---|---|---|
| `main` | Integration branch — all PRs merge here. Treated as unstable/nightly. | No |
| `release/x.y.x` | Stable line for a minor version. Receives cherry-picked fixes from `main`. | No |
| `vX.Y.Z` tag | Signals a production release. | Yes (`pkg.go.dev` auto-indexes) |

## Developer Certificate of Origin (DCO) sign-off

Every commit must be **signed off**, certifying the
[Developer Certificate of Origin](https://developercertificate.org/), a lightweight statement
that you wrote the contribution or otherwise have the right to submit it under the project's
license. This keeps Floci's licensing clean and unambiguous, and it is **required for a pull
request to be merged**.

Sign off by adding the `-s` flag when you commit:

```bash
git commit -s -m "feat(s3): add multipart upload copy-part support"
```

This appends a `Signed-off-by: Your Name <your@email>` trailer using your configured git
identity. If you forget, you can amend the most recent commit with `git commit --amend -s`, or
sign off a range during an interactive rebase.

### Why the DCO and not a CLA

Floci is built by the community, for the community, and the DCO is how it stays that way. There
is no agreement to sign and no rights to hand over. You certify that the work is yours to give,
you keep the copyright in it, and it reaches everyone else on the same MIT terms it arrived
under.

A CLA would ask every contributor to grant something extra to whoever holds the project. Floci
does not ask for that. The Lead Maintainer signs off the same way a first-time contributor
does, and holds no rights over your work that you do not hold over theirs. Code released under
MIT stays under MIT: free to use, fork, and build on, for anyone, permanently.

Changes to this policy are reserved to the Lead Maintainer under
[GOVERNANCE.md](https://github.com/floci-io/.github/blob/main/GOVERNANCE.md).

## Commit Message Format

This project uses [Conventional Commits](https://www.conventionalcommits.org/) — semantic-release reads these to generate the changelog and version bumps automatically.

> **The PR title is validated automatically by CI** and must follow this format, since it becomes the squash-merge commit message that semantic-release reads.

### Format

```
<type>[optional scope]: <description>
```

- **type** — one of the values in the table below (lowercase)
- **scope** — optional, in parentheses, identifies the area (e.g. `container`, `services`, `examples`)
- **description** — short summary in the imperative mood, no trailing period
- Append `!` before the colon to signal a breaking change: `feat(container)!:`

| Type | When to use | Version bump |
|------|-------------|--------------|
| `feat` | New service config, container option, or API | minor |
| `fix` | Bug fix or compatibility correction | patch |
| `perf` | Performance improvement | patch |
| `revert` | Reverts a previous commit | patch |
| `docs` | Documentation only | none |
| `style` | Formatting, whitespace — no logic change | none |
| `chore` | Build, CI, dependencies, housekeeping | none |
| `refactor` | Code restructure without behavior change | none |
| `test` | Adding or updating tests | none |
| `build` | Build system or tooling changes | none |
| `ci` | CI workflow changes | none |
| `BREAKING CHANGE` | Footer or `!` suffix — incompatible change | major |

### Valid examples ✅

```
feat(services): add ElastiCache service config
fix(container): correct default startup timeout
perf(container): reduce wait strategy polling interval
chore: release 1.0.1
docs: update README with credential helpers
refactor(services): extract common service config pattern
test(s3): add bucket creation round-trip test
feat!: remove deprecated WithImage option
ci: add conventional commits lint workflow
build: bump testcontainers-go to v0.43.0
```

### Invalid examples ❌

```
Add S3 support                   # missing type
Feature: add something           # "Feature" is not a valid type
feat : space before colon        # space before colon
feat(services)add missing colon  # missing colon
FIX(s3): uppercase type          # type must be lowercase
feat(my scope): scope has spaces # scope cannot contain spaces
fix(): empty scope               # empty scope
feat(s3):no space after colon    # missing space after colon
wip: still working on this       # "wip" is not a recognised type
```

**No AI attribution.** Do not add "Generated by", "Co-Authored-By: …-bot", or similar trailers to commit messages. Attribution should be limited to human contributors.

## Adding a New Service Configuration

Port from the Java reference module (`testcontainers-floci`): read its `<Service>Config.java` and
`<Service>ConfigTest` and match the env-var keys and defaults exactly.

1. In `services.go`, add a `<Service>Config` struct, a `Default<Service>Config()` constructor and its
   `applyEnvVars` method (and `applyExposedPorts` if the service publishes ports).
2. In `floci.go`, add the field to `FlociContainer`, set its default in `newBuilder()`, add the
   `With<Service>Config` builder method, and call it from `applyAllConfigs` (and `refreshExposedPorts`).
3. In `options.go`, add the package-level `With<Service>Config` option that wraps the builder method, so
   it can be passed to `Run`.
4. If the service spawns sibling containers, add it to `dockerSocketServices` in `services.go` (mark it
   `mockable` if the Java config's `requiresDockerSocket()` checks `!mock`), with a case in
   `socket_internal_test.go`.
5. Add an example under `examples/<service>/` and integration coverage in `floci_test.go`, using
   `floci.Run` and `testcontainers.CleanupContainer`.

## Pull Request Guidelines

1. Branch off `main`: `git checkout -b feature/my-feature`
2. Open a PR targeting `main`.
3. CI runs tests automatically — all checks must pass before merge.
4. Keep PRs focused — one feature or fix per PR.
5. Reference any related issues in the PR description.

### Pull Request Limits and Review Bandwidth

To make sure every contribution gets a thorough, high-quality review in a reasonable time, we ask contributors to keep **no more than 2 open, non-draft pull requests** at any time in this repository.

- **Why this policy exists:** maintainer review time is limited. Capping concurrent open PRs prevents review backlogs, reduces context switching, and keeps PR cycle times short for everyone.
- **Dependent work:** if your work depends on a PR that has not been merged yet, build on that branch or note the dependency in the discussion instead of opening separate, uncoordinated PRs.
- **Draft PRs:** drafts do not count against the limit. Mark a draft as ready for review only when you have review capacity available.
- **How it is applied:** a bot labels your 3rd and later open pull requests `over-pr-limit` with a reminder. Starting **2026-10-08**, from your 5th open pull request onward, new ones are closed automatically. Your branch and commits are kept, and you can reopen the PR once one of your other PRs is merged or closed. Maintainers and dependency bots are not counted.

Once your current pull requests are reviewed, merged, or closed, you are welcome to open new ones!

## Testing Policy for Pull Requests

- Pull requests that introduce new behavior must include tests that validate that behavior.
- Pull requests that fix bugs should include a regression test whenever the bug can be covered realistically.
- Pull requests that do not change observable behavior (docs, formatting, dependency housekeeping) may not require new tests.
- Even when no new tests are needed, the existing test suite must still pass.

If a pull request does not include new tests, the author should explain why in the PR description.

CI runs automatically on every pull request, and all checks must pass before merge.

## Release Process (maintainers)

### New minor or major release

```bash
# 1. Create a release branch from main
git checkout main && git pull
git checkout -b release/1.2.x

# 2. Push — the semver workflow runs semantic-release automatically,
#    bumps the version, updates CHANGELOG.md, and pushes tag v1.2.0.
git push origin release/1.2.x
```

### Patch release on an existing line

```bash
git checkout release/1.1.x
git cherry-pick <commit-sha>
git push origin release/1.1.x
```

### Hotfix

1. Fix on `main` via the normal PR process.
2. Cherry-pick the merge commit onto the relevant `release/x.y.x` branch and push.

## Reporting Security Issues

Please do **not** open public issues for security vulnerabilities. See [SECURITY.md](SECURITY.md) for how to report them privately.
