package tui

const (
	appTitle    = "PortWatch"
	appSubtitle = "\nPort & Process Manager\n"
)

const (
	headerVersionPaddingFmt = "                                      v%s"
	headerNotUpdated        = "not updated"
	headerUpdatedPrefix     = "Updated "
	headerScopePortFmt      = "%s · PORT %d"
	headerNoSelection       = "no selection"
	headerSelectedFmt       = "selected %d/%d"
	headerSummaryFmt        = "\n%s · TCP                         %d results · %s · %s\n"
)

const (
	tableHeaderWide   = "PORT   PROTOCOL   STATE        PID      PROCESS\n"
	tableHeaderNarrow = "PORT   PROTOCOL   PID      PROCESS\n"
	rowMarkerNormal   = "  "
	rowMarkerSelected = "> "
	tableRowWideFmt   = "%s%-5d %-10s %-12s %-8d %s\n"
	tableRowNarrowFmt = "%s%-5d %-10s %-8d %s\n"
)

const (
	emptyNoMatchFmt  = "No match for %q.\n"
	emptyPortFreeFmt = "Port %d is available.\n"
	emptyNoListeners = "No listening ports found.\n"
	selectionNone    = "Selected: -\n"
	selectionFmt     = "Selected: %d · PID %d · %s\n"
)

const (
	filterPromptEditingFmt = "\nFilter: %s_\n"
	filterPromptIdleFmt    = "\nFilter: %s\n"
	statusLineFmt          = "Status: %s\n"
	statusLineIndentedFmt  = "\nStatus: %s\n"
	errorLineFmt           = "\nError: %s\n"
)

const (
	detailsTitle        = "\nProcess Details\n"
	detailsRowFmt       = "%-20s %s\n"
	lookupNoticeFmt     = "\n%s\n"
	detailsActions      = "\nEsc Back   K Kill   R Refresh   T Auto-refresh   Q Quit\n"
	detailsPageLabel    = "Process details"
	fieldPort           = "Port"
	fieldProtocol       = "Protocol"
	fieldState          = "State"
	fieldLocalAddress   = "Local Address"
	fieldRemoteAddress  = "Remote Address"
	fieldPID            = "PID"
	fieldProcessName    = "Process Name"
	fieldParentChain    = "Parent Chain"
	fieldExecutablePath = "Executable Path"
	fieldCommandLine    = "Command Line"
	fieldWorkingDir     = "Working Directory"
)

const (
	confirmTitleFmt = "\nTerminate process?\n\nPID      %d\nProcess  %s\nPort     %d\n"
	confirmWarning  = "\nThis will terminate the process.\n"
	confirmActions  = "\nEnter Confirm   Esc Cancel\n"
)

const (
	helpPageTitle  = "\nHow to use PortWatch\n\n"
	helpStep1      = "1. Select a listening port with Up/Down.\n"
	helpStep2      = "2. Press Enter to inspect the process.\n"
	helpStep3      = "3. Press K, then Enter, to terminate after confirmation.\n"
	helpStep4      = "4. PortWatch verifies the PID and port release.\n\n"
	helpKeysTitle  = "Keys\n"
	helpKeysList   = "↑↓ Select   Enter Details   / Filter   K Kill\n"
	helpKeysFull   = "R Refresh   T Auto-refresh   V Views   ? Help   Esc Back   Q Quit\n"
	helpFooterTop  = "\n↑↓ Select (u/j)   Enter Details   / Filter   K Kill\n"
	helpFooterRest = "R Refresh   T Auto-refresh   V Views   ? Help   Q Quit\n"
)

const (
	viewMenuTitle        = "\nChoose a view\n\n"
	viewMenuRowFmt       = "%s [%s] %s\n"
	viewMenuActions      = "\nL/C/A Select   Esc Back\n"
	markerActive         = ">"
	markerInactive       = " "
	viewOptionListening  = "Listening ports"
	viewOptionConnection = "Active connections"
	viewOptionAll        = "All TCP records"
)

const (
	statusRefreshing       = "Refreshing..."
	statusAutoRefreshOnFmt = "Auto-refresh on, scanning every %v. Press T to stop."
	statusAutoRefreshOff   = "Auto-refresh off."
	statusLoadingFmt       = "Loading %s..."
	statusRefreshFailedFmt = "Refresh failed: %v"
	statusKillFailedFmt    = "Kill failed: %v"
	statusKilledFmt        = "Process terminated. Port %d is available."
	statusVerifying        = "Verifying PID and port release..."
)

const (
	placeholderUnknown = "Unknown"
	placeholderDash    = "-"
	ellipsis           = "..."
	ageSecondsFmt      = "%ds ago"
)

const (
	lookupAccessDeniedText = "Process information unavailable. Access denied."
	lookupExitedText       = "Process information unavailable. Process exited."
	lookupUnavailableText  = "Process information unavailable."

	scopeListeningText   = "LISTENING"
	scopeConnectionsText = "CONNECTIONS"
	scopeAllText         = "ALL"
	scopeUnsupportedFmt  = "%s view is unavailable on this platform: %w"
)
