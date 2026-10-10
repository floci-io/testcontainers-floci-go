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

Releases are cut from `main` by release-please; there are no release branches.

| Branch / tag | Purpose | Published? |
|---|---|---|
| `main` | Integration branch: all PRs merge here. Merging does not release anything by itself. | No |
| release PR (`release-please--branches--main`) | Kept up to date by release-please with the next version and `CHANGELOG.md`. | When a maintainer merges it |
| `vX.Y.Z` tag | A release, created when the release PR is merged. | Yes (Go module proxy, `pkg.go.dev` auto-indexes) |

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

This project uses [Conventional Commits](https://www.conventionalcommits.org/): release-please reads them to compute the next version and write the changelog.

> **Commits are validated by CI** (`commit-lint`). Use the same format for the PR title too: it becomes the squash-merge commit message that release-please reads.

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

The module is laid out as a shared core plus one package per cloud: `internal/core` (the cloud
descriptor, request building, Docker socket detection, the started container with `Reset` and
`Terminate`) and `flociaws` (AWS). The root package `floci` only holds deprecated aliases of `flociaws`;
never add code there, but do add an alias there for every new exported name in `flociaws`.

1. In `flociaws/services.go`, add a `<Service>Config` struct, a `Default<Service>Config()` constructor
   and its `applyEnvVars` method (and `applyExposedPorts` if the service publishes ports).
2. In `flociaws/floci.go`, add the field to `FlociContainer`, set its default in `newBuilder()`, add the
   `With<Service>Config` builder method, and call it from `applyAllConfigs` (and `refreshExposedPorts`).
3. In `flociaws/options.go`, add the package-level `With<Service>Config` option that wraps the builder
   method, so it can be passed to `Run`.
4. If the service spawns sibling containers, add it to `SocketServices` in the `awsDescriptor`
   (`flociaws/floci.go`; `Mockable: true` if the Java config's `requiresDockerSocket()` checks `!mock`),
   with a case in `flociaws/socket_internal_test.go`.
5. Add the deprecated aliases (type, `Default…Config`, `With…Config`) to the root `floci.go`.
6. Add an example under `examples/<service>/` and integration coverage in `flociaws/floci_test.go`,
   using `flociaws.Run` and `testcontainers.CleanupContainer`.

## Pull Request Guidelines

1. Branch off `main`: `git checkout -b feature/my-feature`
2. Open a PR targeting `main`.
3. CI runs tests automatically — all checks must pass before merge.
4. Keep PRs focused — one feature or fix per PR.
5. Reference any related issues in the PR description.

### Pull Request Limits and Review Bandwidth

To make sure every contribution gets a thorough, high-quality review in a reasonable time, we ask contributors to keep **no more than 4 open pull requests**, drafts included, at any time in this repository, and **no more than 2 of them ready for review**.

- **Why this policy exists:** maintainer review time is limited. Capping concurrent open PRs prevents review backlogs, reduces context switching, and keeps PR cycle times short for everyone.
- **Dependent work:** if your work depends on a PR that has not been merged yet, build on that branch or note the dependency in the discussion instead of opening separate, uncoordinated PRs.
- **Draft PRs:** drafts do not count against the limit of 2 ready pull requests, but they do count toward the total of 4. You can have, for example, 2 ready and 2 drafts, or 1 ready and 3 drafts. Use drafts for work in progress, not as a queue of finished changes waiting for a review slot, and mark a draft as ready for review only when you have review capacity available.
- **How it is applied:** a bot turns a pull request back into a draft if it would be your 3rd ready for review, and closes a pull request opened while you already have 4 open, drafts included. Your branch and commits are always kept: mark the draft ready again once one of your ready pull requests is merged, closed or turned into a draft, and reopen a closed pull request once you have fewer than 4 open. Maintainers and dependency bots are not counted.

Once your current pull requests are reviewed, merged, or closed, you are welcome to open new ones!

## Testing Policy for Pull Requests

- Pull requests that introduce new behavior must include tests that validate that behavior.
- Pull requests that fix bugs should include a regression test whenever the bug can be covered realistically.
- Pull requests that do not change observable behavior (docs, formatting, dependency housekeeping) may not require new tests.
- Even when no new tests are needed, the existing test suite must still pass.

If a pull request does not include new tests, the author should explain why in the PR description.

CI runs automatically on every pull request, and all checks must pass before merge.

## Release Process (maintainers)

Releases follow the same release-please flow as the Java module (`.github/workflows/release-please.yml`):

1. Every push to `main` lets release-please update one **release PR**. It computes the next version from
   the Conventional Commits since the last release (`fix:`/`perf:` a patch, `feat:` a minor, `feat!:` or a
   `BREAKING CHANGE:` footer a major; `docs:`, `chore:`, `ci:`, `test:`, `refactor:` none) and writes the
   `CHANGELOG.md` entry. Commit messages are linted in CI (`commit-lint`), because they become the changelog.
2. To release, review and merge the release PR. release-please then tags `vX.Y.Z` and creates the GitHub
   release, and the same workflow confirms the version on the Go module proxy and runs the security scans
   (CodeQL, Trivy, dependency snapshot) against the tag.
3. To recover a release (for example a failed proxy check), run **Release Please** manually with the `tag`
   input set to the existing tag.

Nothing is cherry-picked: fixes merge to `main` and ship with the next release PR.

## Reporting Security Issues

Please do **not** open public issues for security vulnerabilities. See [SECURITY.md](SECURITY.md) for how to report them privately.
