# Import Verification: 2026-09-17

## Scope

Import the owner's supplied new-site source into `jzg-lab/sub2api`.
No production server, account, database, DNS, or live deployment is changed.

## Initial Transfer Checks

- `git ls-remote` and a fresh clone both confirmed an empty target repository.
- 4,106 selected source/documentation files (53,366,229 bytes) were copied
  into an isolated checkout and SHA-256 verified against the original.
- Restored the non-secret `openspec/config.yaml` project metadata separately;
  it was initially excluded by the generic private-config filename filter.
- Excluded local dependencies, compiled frontend output, operational evidence,
  presentation artifacts, archives, private config patterns, and local manifests.
- Reviewed high-signal secret scan findings: placeholder private keys and
  credentials in tests/examples, CI-only database credentials, and existing
  OAuth application constants. No live deployment credential was identified
  by these checks. This is not a comprehensive security guarantee.
- Preserved the upstream license, attribution, application code, lockfile,
  migration files, and Go import namespace.
- Corrected ignore rules so project docs and the admin CLI script are tracked.
- Added a placeholder for the frontend embed directory, whose compiled assets
  are intentionally not committed.
- Installed 973 frontend packages using `pnpm@9.15.9` and
  `--frozen-lockfile`; the supplied lockfile was not changed.
- Passed the production frontend build, including `vue-tsc -b` and the
  3-test locale-key completeness suite.
- Passed the full frontend `lint:check` command.
- Passed the targeted account-editing and traffic-control suites:
  2 files, 62 tests, no failures.
- Build warnings remain for old Browserslist data, mixed static/dynamic
  imports, and chunks exceeding 500 kB. No unrelated dependency upgrades
  or UI/code refactors were performed.
- Shell scripts are recorded with executable Git modes for Linux builds.
- Final staged inventory: 4,110 files. Working-tree SHA-256 comparison found
  4,103 files byte-identical to the delivery, four intended documentation/ignore
  edits, and three new documentation/placeholder files. No business source or
  lockfile changes were found. Git applies the existing LF rules when storing
  text; this does not change source behavior.
- The final inventory contains none of the excluded dependency, verification,
  private environment, archive, binary, or generated frontend-output paths.
- Whitespace checks passed for the newly added maintenance documents,
  ignore-rule additions, and embed placeholder.

## Initial Verification Limits

- No Go executable was found in PATH or the checked conventional local paths.
- The local Docker engine was not running, so backend compilation, database
  integration tests, and container build/start verification were not performed.
- Historical test results supplied with the source are historical evidence,
  not tests performed as part of this import.
- Real upstream requests, payment flows, email, production migrations, and
  live traffic cutover are outside this repository-transfer task.
- The system-default pnpm 11 tried to reconcile the pnpm 9 dependency
  layout and stopped without a TTY. Verification was rerun successfully
  using pnpm 9.15.9; use pnpm 9 for this source tree.
- Obsidian CLI was attempted but reported that Obsidian was not running.
  This import's durable project record is therefore in these docs and Git.

## Publication

The intended default branch is `main`. The initial non-force push was rejected
by GitHub with `GH013` push protection. It detected embedded Google OAuth client
IDs and secrets in:

- `backend/internal/pkg/antigravity/oauth.go`
- `backend/internal/pkg/antigravity/oauth_test.go`
- `backend/internal/pkg/geminicli/constants.go`

No protection bypass, credential obfuscation, or remote policy change was
attempted. These are application constants already present in the supplied
source, not credentials added during import. Their presence does not make
them exempt from the destination repository's security rules.

The owner subsequently approved moving these values to private runtime
configuration and cleaning the unpublished history. The follow-up:

- Removes all four detected credential values, including test assertions.
- Reads both provider client IDs and secrets from environment variables, without
  defaults or obfuscated fallbacks.
- Rejects missing/blank configuration before authorization and token requests.
- Preserves explicit AI Studio custom configuration and the CLI client's
  existing redirect/scope behavior.
- Updates affected tests to use dummy fixtures, adds missing-config and
  precedence coverage, and passes the four variables through Compose.
- Adds blank optional variables to the example and private-env initializer.
- Documents that existing refresh tokens require their original client pair.

Follow-up verification:

- Downloaded Go 1.27.1 from `go.dev`, verified its SHA-256 against the official
  release listing, and extracted it into a temporary local tools directory.
- `go test -tags=unit ./internal/pkg/antigravity ./internal/pkg/geminicli` passed.
- `go test -tags=unit ./internal/service -run
  'Test(AntigravityOAuthService|GeminiOAuthService)' -count=1 -timeout=10m`
  passed. The first attempt could not fetch modules from `proxy.golang.org`;
  retrying via the project's existing `goproxy.cn` mirror succeeded, retaining
  Go checksum verification.
- Compose parsed successfully with dummy values; all four OAuth variables
  reached the app service. No Docker daemon or live container was used.
- `gofmt` and `git diff --check` passed for the changes.
- Full backend build with `go build -tags=embed ... ./cmd/server` passed,
  embedding the previously verified frontend. The executable was written
  outside the repository and was not started. This supersedes the initial
  transfer's missing-Go limitation, but not its live-deployment limitations.

The outgoing branch was rebuilt as a sanitized root commit. An exact-match scan
against the four original credential literals verified they are absent from
the outgoing commit, and neither rejected commit is an ancestor. Generic secret
pattern hits remaining in log-redaction tests and deployment examples were
reviewed as dummy placeholders, not the rejected credential values.

Normal push succeeded on 2026-09-17 without a push-protection bypass or force
push. GitHub's default branch and `refs/heads/main` were read back and verified
as `45c065cd8c535f73527f1aff98e372d6ad8c91e0`, matching the local source baseline.
This publication note is a subsequent documentation-only commit.

Final combined OAuth regression passed for both provider packages and the
service package. The working tree was clean at source publication. No remote
history was overwritten; the target repository was empty. Original local
delivery files remain untouched and are not an automatic deployment source.

Production was not deployed. Configure the private OAuth pairs and validate
existing-account refresh behavior before deploying. A published source baseline
must not be represented as a fully tested production release; GitHub CI results
and real production integrations have not been independently verified here.
