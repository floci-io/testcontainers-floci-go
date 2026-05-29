# Contributing to testcontainers-floci-go

Thank you for your interest in contributing! This is a community-driven Testcontainers module for Floci and all contributions are welcome.

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
go test -run TestS3     # run a single test
```

## Branching Model

This module follows a **tag-driven release model**.

| Branch | Purpose | Published? |
|---|---|---|
| `main` | Integration branch — all PRs merge here. Treated as unstable/nightly. | No |
| `release/x.y.x` | Stable line for a minor version. Receives cherry-picked fixes from `main`. | No |
| `X.Y.Z` tag | Signals a production release. | Yes (`pkg.go.dev` auto-indexes) |

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

Do not include `Co-Authored-By` trailers for AI tools in commit messages. Attribution should be limited to human contributors.

## Adding a New Service Configuration

1. Add a `<Service>Config` struct and a `Default<Service>Config()` constructor in `services.go`
2. Add the corresponding field to `FlociOptions` in `floci.go`
3. Wire the config into the container environment variables in the `applyOptions` function
4. Add an example under `examples/<service>/`
5. Add integration tests in `floci_test.go`

## Pull Request Guidelines

1. Branch off `main`: `git checkout -b feature/my-feature`
2. Open a PR targeting `main`.
3. CI runs tests automatically — all checks must pass before merge.
4. Keep PRs focused — one feature or fix per PR.
5. Reference any related issues in the PR description.

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
#    bumps the version, updates CHANGELOG.md, and pushes tag 1.2.0.
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

Please do **not** open public issues for security vulnerabilities. Report them privately by emailing the maintainer or using [GitHub private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing/privately-reporting-a-security-vulnerability).
