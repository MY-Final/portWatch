# PortWatch

PortWatch is a Windows developer CLI for finding TCP listening ports and the
processes that own them.

## MVP

Build on Windows with Go 1.22 or newer:

```powershell
go build -o portwatch.exe ./cmd/portwatch
```

Inspect one port:

```powershell
.\portwatch.exe 8080
```

List all TCP listeners:

```powershell
.\portwatch.exe
```

Safely release a port. PortWatch shows the process details, asks for explicit
confirmation, terminates the process, and scans again to verify release:

```powershell
.\portwatch.exe free 8080
```

Process details that require elevated access may be unavailable. Run the
terminal as Administrator when Windows reports access denied. The current MVP
supports Windows TCP LISTENING only; JSON, watch, TUI, Linux, and macOS are
planned for later phases.

## Development

```powershell
go test ./...
go build ./...
```

CI tests Windows, Linux, and macOS. Release artifacts are configured through
GoReleaser for amd64 and arm64 on all three platforms.
