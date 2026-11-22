#!/usr/bin/env bash
# Release helper for the gracewrap Go module. See RELEASING.md.
#
#   scripts/release.sh prepare vX.Y.Z   # branch + move CHANGELOG "Unreleased" into a version section
#   scripts/release.sh tag     vX.Y.Z   # on an up-to-date main: test, tag and push (triggers the Release workflow)
#   scripts/release.sh check   vX.Y.Z   # validate the version string + CHANGELOG section (used by CI)
#   scripts/release.sh notes   vX.Y.Z   # print the release notes for a version (used by CI)
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
CHANGELOG=CHANGELOG.md

die() { echo "error: $*" >&2; exit 1; }

normalize() {
  local v=${1:-}
  [[ -n "$v" ]] || die "missing version (e.g. v0.2.0)"
  [[ "$v" == v* ]] || v="v$v"
  printf '%s' "$v"
}

# Body of the "## [X.Y.Z]" section of CHANGELOG.md (empty if there is none).
section() {
  local ver=${1#v}
  # Collect the lines between this header and the next "## [", trimming
  # leading and trailing blank lines.
  awk -v hdr="## [$ver]" '
    index($0, hdr) == 1 { on = 1; next }
    on && /^## \[/      { exit }
    on                  { buf[++n] = $0 }
    END {
      s = 1; while (s <= n && buf[s] ~ /^[[:space:]]*$/) s++
      e = n; while (e >= s && buf[e] ~ /^[[:space:]]*$/) e--
      for (i = s; i <= e; i++) print buf[i]
    }
  ' "$CHANGELOG"
}

cmd_check() {
  local tag; tag=$(normalize "${1:-}")
  local semver='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$'
  [[ "$tag" =~ $semver ]] || die "'$tag' is not a semantic version like v1.2.3 or v1.2.3-rc.1"

  # Go requires a /vN suffix on the module path for major versions >= 2.
  local major=${BASH_REMATCH[1]} mod
  mod=$(awk '/^module /{print $2; exit}' go.mod)
  if (( major >= 2 )) && [[ "$mod" != */v"$major" ]]; then
    die "$tag is major version $major, but go.mod's module path is '$mod'. Go needs it to end in /v$major (see RELEASING.md)."
  fi

  [[ -n "$(section "$tag")" ]] || die "$CHANGELOG has no '## [${tag#v}]' section with content. Run: scripts/release.sh prepare $tag"
  echo "ok: $tag"
}

cmd_notes() {
  local tag; tag=$(normalize "${1:-}")
  local mod; mod=$(awk '/^module /{print $2; exit}' go.mod)
  section "$tag"
  cat <<EOF

---

**Install**

\`\`\`bash
go get ${mod}@${tag}
\`\`\`

Docs: https://pkg.go.dev/${mod}@${tag}
EOF
}

cmd_prepare() {
  local tag; tag=$(normalize "${1:-}")
  local ver=${tag#v} today; today=$(date -u +%Y-%m-%d)

  [[ -z "$(git status --porcelain)" ]] || die "working tree is not clean"
  grep -q '^## \[Unreleased\]' "$CHANGELOG" || die "$CHANGELOG has no '## [Unreleased]' section"
  grep -q "^## \[$ver\]" "$CHANGELOG" && die "$CHANGELOG already has a section for $ver"
  [[ -n "$(section Unreleased)" ]] || die "'## [Unreleased]' in $CHANGELOG is empty; nothing to release"

  git fetch -q origin main
  git switch --no-track -c "release/$tag" origin/main

  # Insert the version header right below "## [Unreleased]", so everything that
  # was unreleased now belongs to this version and Unreleased starts out empty.
  awk -v hdr="## [$ver] - $today" '
    { print }
    /^## \[Unreleased\]/ && !done { print ""; print hdr; done = 1 }
  ' "$CHANGELOG" > "$CHANGELOG.tmp" && mv "$CHANGELOG.tmp" "$CHANGELOG"

  cmd_check "$tag" >/dev/null
  git add "$CHANGELOG"
  git commit -q -m "chore(release): $tag"

  cat <<EOF
Prepared release/$tag. Release notes will be:
-----------------------------------------------
$(section "$tag")
-----------------------------------------------
Next:
  git push -u origin release/$tag   # then open a PR and merge it
  make release VERSION=$tag         # from an up-to-date main, after the merge
EOF
}

cmd_tag() {
  local tag; tag=$(normalize "${1:-}")

  [[ "$(git rev-parse --abbrev-ref HEAD)" == main ]] || die "switch to main first (releases are cut from main)"
  [[ -z "$(git status --porcelain)" ]] || die "working tree is not clean"
  git fetch -q --tags origin main
  [[ "$(git rev-parse HEAD)" == "$(git rev-parse origin/main)" ]] || die "local main differs from origin/main. Run: git pull --ff-only"
  git rev-parse -q --verify "refs/tags/$tag" >/dev/null && die "tag $tag already exists"

  cmd_check "$tag"
  echo "running tests..."
  go vet ./...
  go test -race ./...

  git tag -a "$tag" -m "gracewrap $tag"
  git push origin "$tag"
  echo "Pushed $tag. The Release workflow now publishes it:"
  echo "  https://github.com/imran31415/gracewrap/actions/workflows/release.yml"
}

case "${1:-}" in
  check)   shift; cmd_check "$@" ;;
  notes)   shift; cmd_notes "$@" ;;
  prepare) shift; cmd_prepare "$@" ;;
  tag)     shift; cmd_tag "$@" ;;
  *) sed -n '2,7p' "$0" | sed 's/^# \{0,1\}//'; exit 2 ;;
esac
