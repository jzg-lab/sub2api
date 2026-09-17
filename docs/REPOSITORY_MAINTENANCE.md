# Repository Maintenance

## Source Of Truth

- Business source repository: `https://github.com/jzg-lab/sub2api.git`.
- Initial import: the owner's local new-site delivery, reviewed on 2026-09-17.
- Original project: `https://github.com/Wei-Shaw/sub2api`. Preserve its license
  and attribution. The existing Go module name is intentionally unchanged;
  it is an import namespace, not the deployment source.
- This repository was empty when imported. There was no remote branch,
  commit history, or old release to overwrite or back up.
- The import is a source baseline, not a production deployment or proof of
  production readiness. Historical source documents describe previous work;
  their test and server status claims were not rerun or reconfirmed here.

## Build The Custom Source

Use a separate checkout and private deployment configuration. Do not reuse
the upstream `weishaw/sub2api` image or upstream installation script for this
custom distribution. The default `deploy/docker-compose.yml` is an inherited
upstream-image template; the custom source-build entry is
`deploy/compose.relay.yml`.

For a NEW isolated installation on a Docker/Compose host:

```sh
git clone https://github.com/jzg-lab/sub2api.git
cd sub2api
git log -1 --oneline
cd deploy
python3 init-relay-env.py --bind 127.0.0.1
docker compose --env-file .env -f compose.relay.yml config --quiet
docker compose --env-file .env -f compose.relay.yml build app
docker compose --env-file .env -f compose.relay.yml up -d --wait
python3 smoke-relay.py --env .env
```

These commands initialize a separate database and are NOT an existing-business
migration procedure. They are documented only; the import did not run them.
Do not regenerate an existing `.env`, change database volumes, or overwrite
JWT/TOTP keys during an upgrade.

For existing business deployments, first confirm the active Compose project,
database/schema, volumes, secrets, reverse proxy, and build source. Back up
the database and configuration, retain the old image and source commit, and
test migrations and real upstream calls in isolation before switching.
Never use `docker compose down -v` as an update command.

## Updates And Releases

Pull from this repository, then explicitly rebuild the custom source.
Record the full commit and built image digest for every deployment. Prefer
an approved, pinned commit over an unreviewed moving branch.

The inherited release workflow runs on `v*` tags or manual dispatch. This
import does not create a `v*` tag, trigger a release intentionally, configure
registry credentials, or claim that a custom GHCR image already exists.
The original application's in-place updater is already disabled by this
distribution's constructor; do not re-enable upstream replacement updates.

Do not force-push to overwrite shared work. Fetch before publishing changes.
If another contributor has updated the branch, integrate and verify their
changes before pushing. Keep project decisions and verification in `docs/`
and Git commits.

## Private OAuth Configuration

Google OAuth client IDs and secrets are no longer embedded in this distribution.
For Antigravity or Gemini CLI / Google One OAuth, provide the corresponding pair
in the server's private `deploy/.env` (or its secret-management system):

```dotenv
ANTIGRAVITY_OAUTH_CLIENT_ID=
ANTIGRAVITY_OAUTH_CLIENT_SECRET=
GEMINI_CLI_OAUTH_CLIENT_ID=
GEMINI_CLI_OAUTH_CLIENT_SECRET=
```

The supplied source-build Compose file passes these variables to the application.
Native deployments must export them into the application's process environment;
a `.env` file is not loaded automatically by the Go binary. Restart/recreate the
application after changing its environment.

Use only credentials you are authorized to use, with the redirect URIs and
scopes required by the integration. Existing refresh tokens are bound to their
original OAuth client: preserve the matching pair or reauthorize the accounts
using the intended client. Merely inventing a client ID or switching to an
unrelated Google OAuth app does not preserve existing refresh-token behavior.

Both fields are required for the corresponding provider's authorization,
token exchange, and refresh. Missing or whitespace-only values cause explicit
configuration errors before token HTTP requests. There is no built-in fallback.
Other providers and normal startup do not require these optional pairs.

The separate AI Studio custom OAuth configuration remains supported. It is not
the same as `GEMINI_CLI_OAUTH_*`; do not replace the existing AI Studio settings
with the CLI pair.

No credentials were copied into a new private configuration by this import, and
no production environment was changed. Configure and validate the pairs before
deploying this version if existing accounts depend on them. Never commit `.env`,
provider tokens, or a filled-in example file.

## Import Hygiene

The initial import excludes local dependencies, build outputs, verification
logs/screenshots, presentation artifacts, archives, private environment files,
runtime databases, and local delivery manifests. The original delivery is
left unchanged outside this checkout.

The retained source was checked for common token/private-key patterns.
Reviewed matches include dummy test fixtures, example credentials, and
inherited OAuth application constants; a pattern scan is not a security audit.

The initial transfer changed only repository documentation, ignore rules,
executable shell-script modes, and the missing Go embed placeholder. Following
GitHub push-protection rejection, the owner approved removal of the embedded
OAuth credentials and migration to private environment configuration. That
follow-up changes the two providers' OAuth configuration requirements and adds
tests and deployment passthrough. Billing, database migrations, and UI behavior
are unchanged.

See [import verification](IMPORT_VERIFICATION_2026-09-17.md) for checks and
limitations of this transfer.
