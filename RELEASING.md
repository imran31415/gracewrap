# Releasing gracewrap

gracewrap is a Go module, so a release is a **semver git tag** (`vX.Y.Z`) on `main`.
Once a tag exists, anyone can depend on that exact version:

```bash
go get github.com/imran31415/gracewrap@v0.2.0   # a specific version
go get github.com/imran31415/gracewrap@latest   # newest non-prerelease
```

Pushing the tag starts the [Release workflow](.github/workflows/release.yml), which:

1. validates the version and checks that `CHANGELOG.md` has a section for it,
2. checks the tag is on `main`, then runs `go vet`, the race tests, the example builds and `govulncheck`,
3. creates a GitHub Release whose notes are that CHANGELOG section,
4. fetches the version through `proxy.golang.org`, so it's cached in the Go module proxy and
   listed on [pkg.go.dev](https://pkg.go.dev/github.com/imran31415/gracewrap) right away.

> **Tags are permanent.** The Go module proxy caches a version forever, even if you delete
> the tag afterwards. Never move or re-push a tag. If a release is broken, publish the next
> patch version (and optionally add a `retract` directive to `go.mod`).

## Choosing the version

The project follows [Semantic Versioning](https://semver.org/):

| Change | Bump | Example |
|---|---|---|
| Bug fixes and dependency updates, no API change | patch | `v0.2.0` → `v0.2.1` |
| New API, backwards compatible | minor | `v0.2.1` → `v0.3.0` |
| Breaking API change (or raising the minimum Go version) | major (minor while on `v0`) | `v0.3.0` → `v0.4.0` |
| Release candidate | pre-release suffix | `v0.4.0-rc.1` |

While the version is `v0.x`, the API may still change between minor versions.
`v1.0.0` is a promise of API stability. After that, **a breaking change means `v2`**, and Go
requires the module path to change to `github.com/imran31415/gracewrap/v2` (update `go.mod`
and all imports). The release script refuses a `v2+` tag until the path matches.

## Steps

Record changes under `## [Unreleased]` in `CHANGELOG.md` as PRs land.

### 1. Prepare the release (a PR)

```bash
git switch main && git pull --ff-only
make release-prepare VERSION=v0.2.0
git push -u origin release/v0.2.0
gh pr create --fill --base main
```

`release-prepare` creates a `release/v0.2.0` branch from `origin/main` and moves everything under
`## [Unreleased]` into a new `## [0.2.0] - YYYY-MM-DD` section. Review the notes in the PR, then merge it.

### 2. Tag it

Either of these works.

**From your machine:**

```bash
git switch main && git pull --ff-only
make release VERSION=v0.2.0
```

This re-checks the CHANGELOG, runs the tests, creates an annotated tag and pushes it.

**From GitHub** (no local checkout needed, works from a phone):
Actions → **Release** → **Run workflow** → enter `v0.2.0`. The workflow tags the current tip of `main` itself.

### 3. Check it

- The Release run is green: <https://github.com/imran31415/gracewrap/actions/workflows/release.yml>
- The release is listed: <https://github.com/imran31415/gracewrap/releases>
- `go list -m github.com/imran31415/gracewrap@v0.2.0` resolves from a clean directory.

## Scope

Only the root module (`github.com/imran31415/gracewrap`) is released. `examples/`, `demo/` and
`proof_tests/` are separate modules. They use a `replace` directive to build against the local
checkout, are not tagged, and are left out of the module zip that `go get` downloads.
