# Feature acceptance record

## Verified

- [V] Frontend build, TypeScript, i18n key check and full ESLint completed successfully. See `verification/features/frontend-build-final.log`, `frontend-lint-full.log`, and JSON test reports.
- [V] Backend route/DI/middleware integration tests passed. User cleanup and intelligent test service/handler suites passed; PostgreSQL intelligent queue tests cover migrations, leases, concurrency, history, visibility and group IDOR.
- [V] Migrations 245–247 applied twice to an isolated restored database. 104 pre-existing tables retained matching normalized hashes; password, balance and business fields were preserved. The only intended data change is migration 246's first-admin promotion to `super_admin`.
- [V] Isolated HTTP acceptance passed 304 checks, including asynchronous enqueue and polling, idempotency, protocol-compatible fixture dispatch, protection defaults and confirmed disable, role restrictions, cleanup preview/revalidation/replay/audit, hidden/public result APIs, image CSP, and re-registration after identity archive. Evidence: server `evidence/features/api-smoke.json` (fixture-only, no real model quality claim).
- [V] The strategy registry, dynamic strategy selector and model-selection fixes were deployed as `chengchuan-relay:features-strategies-20260913`. The app container is healthy; the pre-deploy database/app-data backup is under `/opt/chengchuan-relay/backups/before-strategy-deploy-20260913-144600/`, and PostgreSQL/Redis volumes were not recreated.

## Domain and administrator credentials: 2026-09-13

- [V] Public DNS and the server resolver return `107.149.55.60` for `juhe.pub`, with no AAAA record.
- [V] Caddy configuration validation and reload succeeded. A Let's Encrypt certificate covers `juhe.pub`; TLS verification succeeds without disabling certificate validation, and HTTP redirects to HTTPS with status 308.
- [V] HTTPS health, login HTML, entry assets, authenticated account listing and intelligent test settings return 200. The administrator's newly requested credentials successfully authenticate as user 1 with role `super_admin`.
- [V] Credentials were changed using `PUT /api/v1/admin/users/1`. The former JWT returns 401, and audit record 56 redacts the password. Balance, role, status and other administrator fields match the pre-change values.
- [V] Only the requested URL settings were submitted through the settings API; all other settings remain equal in the before/after comparison.
- Evidence: `verification/features/domain-20260913/admin-credentials-verification.json`, `https-verification.json`, and `domain-settings-verification.json`. These are HTTP/TLS checks; no additional browser UI or model-quality result is claimed.

## Browser note

Automated Chromium reached the deployed admin UI and captured account-list and protection-confirmation screenshots. The repository's mandatory onboarding tour overlay prevented the remaining scripted clicks; this is recorded as a tooling limitation, not a product pass. Component and integration tests cover the same interactions. Screenshots are under `verification/features/browser/`.

## Pending external quality check

The server has no real upstream accounts. The local fixture is explicitly labelled `ISOLATED HTTP TEST FIXTURE`; its SVG and answer are only deterministic transport/evaluator inputs. Import a real account before judging model quality, anti-degradation effectiveness or actual upstream billing.

## Anti-degradation strategy update: 2026-09-13

- [V] Source changes add three explicit administrator choices: mode 1 (v3), mode 2 (old scheme), and the original `sub2初代` strategy (`legacy`: session fingerprint, nodejs24 TLS, concurrency cap 16). The effective `protection_mode` is returned by the API and shown together with account identity in the edit dialog.
- [V] Switching strategies restores the current snapshot before applying the selected strategy; ordinary stale account edits preserve legacy/v3 managed fields. Backend service tests and frontend API/UI tests pass (16/16 targeted UI/API tests).
- [V] Built and deployed as `chengchuan-relay:features-strategies-20260913` without changing the database volume. Container health is passing and the in-container health endpoint returns `{"status":"ok"}`; `https://juhe.pub/` returns 200. The public `/health` path is not exposed by the current Caddy route and returns 404.
- [V] Live admin account listing now returns the concrete `protection_mode` for existing accounts; a live legacy preview correctly reports the active mode and safe-switch behavior without mutating the account.
