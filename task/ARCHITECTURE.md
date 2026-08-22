# PortWatch MVP Architecture

## Current repository

`portWatch/` currently contains only `prd/v1.md`; there is no Go module, source package, test suite, or CI configuration. The task plan therefore defines the initial package boundaries instead of adapting an existing implementation.

## Runtime flow

```text
cmd/portwatch/main.go
  -> internal/command.Run
      -> port.Scanner (TCP LISTENING records)
      -> process.Manager (details by PID)
      -> internal/command output/error helpers
```

`free` adds a confirmation gate, calls `Manager.Terminate` for the selected PID, then invokes `Scanner.Port` again. A successful kill is never reported without the second scan showing no matching record.

## Dependency direction

`pkg/model` has no internal dependencies. `internal/port` and `internal/process` depend only on model plus standard library/platform packages. `internal/command` depends on both interfaces and model. `cmd` depends only on command. Platform files implement interfaces and never import command.

## Platform boundary

The interface files are platform-neutral. Implementations are selected at compile time with `//go:build windows`, `linux`, or `darwin`. MVP only ships the Windows implementation. Unsupported platforms must return `ErrUnsupported` from a small tagged stub added by the integration task if needed for cross-compilation; they must not contain runtime OS-name switches.

## Error and safety contract

Library packages return wrapped sentinel errors and never print, panic, or exit. The command package maps errors to stable exit codes. Termination always targets one validated PID, requires explicit interactive confirmation in the command layer, and is followed by process/port verification.

## Parallel ownership

After project initialization and models, scanner interface, process interface, CLI parsing, and error mapping can proceed independently. Windows scanner and process implementation are independent files. The only intentional same-file sequence is `007 -> 008` for `internal/process/windows.go`, followed by `013` for final `main.go` wiring. This is reflected in `INDEX.md`.

