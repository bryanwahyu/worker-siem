# Security Policy

## Reporting a vulnerability

Report it privately through
[GitHub Security Advisories](https://github.com/bryanwahyu/worker-siem/security/advisories/new).
The report stays visible only to the maintainers until a fix is out.

Please do not open a public issue or disclose the problem before it is fixed.
Include what you found, how to reproduce it, and what an attacker could do
with it.

## Supported versions

Only the latest commit on `main` receives fixes.

## Running the worker safely

- Keep the API key out of the config file. Set `MINI_SIEM_API_KEY` in the
  environment, or reference it as `${MINI_SIEM_API_KEY}` in the YAML.
- Use an `https://` base URL. The API key travels in the `X-API-Key` header.
- The spool directory holds raw log events until they are uploaded. The worker
  creates it as `0700` and spool files as `0600`; keep it that way.
- The metrics endpoint has no authentication. Bind `metrics.listen_addr` to
  `127.0.0.1` or a private interface unless you need it reachable.
- `deploy/systemd/mini-siem-worker.service` ships with systemd sandboxing
  enabled. Keep it when you adapt the unit file.
