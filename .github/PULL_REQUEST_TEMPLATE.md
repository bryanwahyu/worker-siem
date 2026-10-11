## What this changes

<!-- One or two sentences. The diff says what; say why. -->

## Related issues

<!-- e.g. Closes #12 -->

## How it was verified

- [ ] `gofmt -l .` prints nothing
- [ ] `go vet ./...`
- [ ] `go test -race ./...`
- [ ] Ran the worker against real logs (say which tailer)

## Notes for the reviewer

- [ ] Changes the spool file format (existing spools still readable?)
- [ ] Changes the event payload sent to the Mini SIEM API
- [ ] Changes a config key or environment variable
