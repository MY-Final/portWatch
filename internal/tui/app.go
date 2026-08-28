package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MY-Final/portWatch/internal/port"
	"github.com/MY-Final/portWatch/internal/process"
	"github.com/MY-Final/portWatch/internal/processinfo"
	"github.com/MY-Final/portWatch/pkg/model"
)

// Model is the Bubble Tea state for the developer-facing port manager.
// Detail and ConfirmKill remain as compatibility fields for callers that used
// the original TUI model; Page is the authoritative view state.
type Model struct {
	Context context.Context
	Scanner port.Scanner
	Manager process.Manager
	Version string

	Scope         port.Scope
	ViewSelection port.Scope
	Page          pageMode
	Ports         []model.PortInfo
	Infos         map[int]model.ProcessInfo
	LookupErrors  map[int]error
	Filter        string
	PortFilter    int
	Filtering     bool
	Width         int
	Height        int
	Err           error
	Status        string
	NextStatus    string
	UpdatedAt     time.Time
	Selected      int
	SelectedKey   rowKey
	DetailRecord  model.PortInfo
	Detail        string
	ConfirmKill   bool
	ConfirmReturn pageMode
	HelpReturn    pageMode
	AutoRefresh   bool
	// refreshInFlight keeps the polling loop from stacking scans when one
	// refresh is slower than autoRefreshInterval.
	refreshInFlight bool
}

// autoRefreshInterval paces the optional polling loop toggled with T.
const autoRefreshInterval = 2 * time.Second

type portsLoadedMsg struct {
	ports        []model.PortInfo
	infos        map[int]model.ProcessInfo
	lookupErrors map[int]error
	scope        port.Scope
	updatedAt    time.Time
}
type portsFailedMsg struct{ err error }
type killDoneMsg struct{ err error }

// autoRefreshTickMsg fires on every autoRefreshInterval while the polling
// loop is enabled; each tick triggers one scan and schedules the next tick.
type autoRefreshTickMsg struct{}

func autoRefreshTick() tea.Cmd {
	return tea.Tick(autoRefreshInterval, func(time.Time) tea.Msg { return autoRefreshTickMsg{} })
}

func New(scanner port.Scanner, manager process.Manager) Model {
	return NewWithPort(scanner, manager, 0)
}

// NewWithPort creates a TUI model optionally focused on one port. The port
// filter is applied locally so the scanner contract remains unchanged.
func NewWithPort(scanner port.Scanner, manager process.Manager, portFilter int) Model {
	return Model{
		Context:       context.Background(),
		Scanner:       scanner,
		Manager:       manager,
		Scope:         port.ScopeListening,
		ViewSelection: port.ScopeListening,
		Page:          pageList,
		PortFilter:    portFilter,
	}
}

func (m Model) Init() tea.Cmd { return m.refresh() }

func (m Model) refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := m.Context
		if ctx == nil {
			ctx = context.Background()
		}
		scope := m.Scope
		if scope > port.ScopeAll {
			scope = port.ScopeListening
		}
		var (
			ports []model.PortInfo
			err   error
		)
		if scoped, ok := m.Scanner.(port.ScopedScanner); ok {
			ports, err = scoped.ListScope(ctx, scope)
		} else if scope == port.ScopeListening {
			if m.Scanner == nil {
				return portsFailedMsg{err: errors.New("port scanner is nil")}
			}
			ports, err = m.Scanner.List(ctx)
		} else {
			return portsFailedMsg{err: unsupportedScopeError(scope)}
		}
		if err != nil {
			return portsFailedMsg{err: err}
		}
		infos := make(map[int]model.ProcessInfo, len(ports))
		infoErrors := make(map[int]error)
		if m.Manager != nil {
			resolvable := make([]model.PortInfo, 0, len(ports))
			for _, record := range ports {
				if record.PID > 0 {
					resolvable = append(resolvable, record)
				}
			}
			infos, infoErrors = processinfo.Resolve(ctx, m.Manager, resolvable)
			processinfo.ApplyNames(ports, infos)
		}
		return portsLoadedMsg{
			ports: ports, infos: infos, lookupErrors: infoErrors,
			scope: scope, updatedAt: time.Now(),
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch value := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(value)
	case tea.WindowSizeMsg:
		m.Width, m.Height = value.Width, value.Height
	case autoRefreshTickMsg:
		if !m.AutoRefresh {
			return m, nil
		}
		if m.refreshInFlight {
			return m, autoRefreshTick()
		}
		m.refreshInFlight = true
		return m, tea.Batch(m.refresh(), autoRefreshTick())
	case portsLoadedMsg:
		previousKey := m.currentKey()
		previousIndex := m.Selected
		m.Ports = append([]model.PortInfo(nil), value.ports...)
		m.Infos = value.infos
		m.LookupErrors = value.lookupErrors
		m.Scope = value.scope
		m.UpdatedAt = value.updatedAt
		m.Err = nil
		m.Status = m.NextStatus
		m.NextStatus = ""
		m.refreshInFlight = false
		sortPorts(m.Ports)
		m.Selected = restoreSelection(m.Ports, previousKey, previousIndex)
		m.SelectedKey = m.currentKey()
	case portsFailedMsg:
		m.NextStatus = ""
		m.Err = value.err
		m.Status = fmt.Sprintf(statusRefreshFailedFmt, value.err)
		m.refreshInFlight = false
	case killDoneMsg:
		if value.err != nil {
			m.ConfirmKill = false
			m.Page = m.ConfirmReturn
			m.Err = value.err
			m.Status = fmt.Sprintf(statusKillFailedFmt, value.err)
			return m, nil
		}
		m.Err = nil
		m.ConfirmKill = false
		m.Page = pageList
		m.NextStatus = fmt.Sprintf(statusKilledFmt, m.DetailRecord.Port)
		m.Status = statusVerifying
		return m, m.refresh()
	}
	return m, nil
}

func (m Model) View() string {
	switch m.Page {
	case pageDetails:
		return m.viewDetails()
	case pageConfirm:
		return m.viewConfirm()
	case pageHelp:
		return m.viewHelp()
	case pageView:
		return m.viewViewMenu()
	default:
		return m.viewList()
	}
}

func (m Model) viewList() string {
	var b strings.Builder
	m.writeHeader(&b)
	if m.Err != nil && len(m.Ports) == 0 {
		fmt.Fprintf(&b, errorLineFmt, cleanError(m.Err))
	} else {
		b.WriteString("\n")
		m.writeTable(&b)
	}
	m.writeSelection(&b)
	if m.Filtering {
		fmt.Fprintf(&b, filterPromptEditingFmt, m.Filter)
	} else {
		fmt.Fprintf(&b, filterPromptIdleFmt, displayFilter(m.Filter))
	}
	if m.Status != "" {
		fmt.Fprintf(&b, statusLineFmt, m.Status)
	}
	m.writeHelp(&b)
	return b.String()
}

func (m Model) writeHeader(b *strings.Builder) {
	b.WriteString(appTitle)
	if m.Version != "" {
		fmt.Fprintf(b, headerVersionPaddingFmt, m.Version)
	}
	b.WriteString(appSubtitle)
	count := len(m.visibleIndexes())
	updated := headerNotUpdated
	if !m.UpdatedAt.IsZero() {
		updated = headerUpdatedPrefix + formatAge(m.UpdatedAt)
	}
	label := scopeLabel(m.Scope)
	if m.PortFilter > 0 {
		label = fmt.Sprintf(headerScopePortFmt, label, m.PortFilter)
	}
	position := headerNoSelection
	if len(m.visibleIndexes()) > 0 {
		position = fmt.Sprintf(headerSelectedFmt, visiblePosition(m.visibleIndexes(), m.Selected)+1, count)
	}
	fmt.Fprintf(b, headerSummaryFmt, label, count, updated, position)
}

func visiblePosition(indexes []int, selected int) int {
	for position, index := range indexes {
		if index == selected {
			return position
		}
	}
	return 0
}

func (m Model) writeTable(b *strings.Builder) {
	if m.Width >= 110 {
		b.WriteString(tableHeaderWide)
	} else {
		b.WriteString(tableHeaderNarrow)
	}
	indexes := m.tableIndexes()
	for _, index := range indexes {
		record := m.Ports[index]
		marker := rowMarkerNormal
		if index == m.Selected {
			marker = rowMarkerSelected
		}
		name := m.processName(record)
		if m.Width >= 110 {
			nameWidth := m.Width - 43
			if nameWidth < 12 {
				nameWidth = 12
			}
			fmt.Fprintf(b, tableRowWideFmt, marker, record.Port, display(record.Protocol), display(record.State), record.PID, fitText(name, nameWidth))
			continue
		}
		nameWidth := m.Width - 32
		if nameWidth < 12 {
			nameWidth = 12
		}
		fmt.Fprintf(b, tableRowNarrowFmt, marker, record.Port, display(record.Protocol), record.PID, fitText(name, nameWidth))
	}
	if len(indexes) == 0 {
		switch {
		case strings.TrimSpace(m.Filter) != "":
			fmt.Fprintf(b, emptyNoMatchFmt, m.Filter)
		case m.PortFilter > 0:
			fmt.Fprintf(b, emptyPortFreeFmt, m.PortFilter)
		default:
			b.WriteString(emptyNoListeners)
		}
	}
}

// tableIndexes keeps the active row inside the terminal viewport. Without a
// bounded window, a long listener list can push the selected row off-screen
// while the summary still reports it as selected.
func (m Model) tableIndexes() []int {
	indexes := m.visibleIndexes()
	if len(indexes) == 0 || m.Height <= 0 {
		return indexes
	}
	limit := m.Height - 10
	if limit < 1 {
		limit = 1
	}
	if len(indexes) <= limit {
		return indexes
	}
	selectedPosition := visiblePosition(indexes, m.Selected)
	start := selectedPosition - limit + 1
	if start < 0 {
		start = 0
	}
	if start+limit > len(indexes) {
		start = len(indexes) - limit
	}
	return indexes[start : start+limit]
}

func (m Model) writeSelection(b *strings.Builder) {
	record, ok := m.selectedRecord()
	if !ok {
		b.WriteString(selectionNone)
		return
	}
	fmt.Fprintf(b, selectionFmt, record.Port, record.PID, m.processName(record))
}

func (m Model) viewDetails() string {
	var b strings.Builder
	m.writeHeader(&b)
	b.WriteString(detailsTitle)
	record := m.DetailRecord
	info := m.Infos[record.PID]
	fields := [][2]string{
		{fieldPort, fmt.Sprint(record.Port)},
		{fieldProtocol, display(record.Protocol)},
		{fieldState, display(record.State)},
		{fieldLocalAddress, display(record.LocalAddr)},
		{fieldRemoteAddress, display(record.RemoteAddr)},
		{fieldPID, fmt.Sprint(record.PID)},
		{fieldProcessName, m.processName(record)},
		{fieldParentChain, display(m.parentChain(info))},
		{fieldExecutablePath, display(info.Executable)},
		{fieldCommandLine, display(info.Command)},
		{fieldWorkingDir, display(info.WorkingDir)},
	}
	for _, field := range fields {
		fmt.Fprintf(&b, detailsRowFmt, field[0], fitText(field[1], maxDetailWidth(m.Width)))
	}
	if err, ok := m.LookupErrors[record.PID]; ok {
		fmt.Fprintf(&b, lookupNoticeFmt, lookupMessage(classifyLookupError(err)))
	}
	if m.Status != "" {
		fmt.Fprintf(&b, statusLineIndentedFmt, m.Status)
	}
	b.WriteString(detailsActions)
	return b.String()
}

func (m Model) viewConfirm() string {
	var b strings.Builder
	m.writeHeader(&b)
	record := m.DetailRecord
	fmt.Fprintf(&b, confirmTitleFmt, record.PID, m.processName(record), record.Port)
	b.WriteString(confirmWarning)
	b.WriteString(confirmActions)
	return b.String()
}

func (m Model) viewHelp() string {
	var b strings.Builder
	m.writeHeader(&b)
	b.WriteString(helpPageTitle)
	b.WriteString(helpStep1)
	b.WriteString(helpStep2)
	b.WriteString(helpStep3)
	b.WriteString(helpStep4)
	b.WriteString(helpKeysTitle)
	b.WriteString(helpKeysList)
	b.WriteString(helpKeysFull)
	return b.String()
}

func (m Model) viewViewMenu() string {
	var b strings.Builder
	m.writeHeader(&b)
	b.WriteString(viewMenuTitle)
	for _, option := range []struct {
		key   string
		label string
		scope port.Scope
	}{
		{key: "L", label: viewOptionListening, scope: port.ScopeListening},
		{key: "C", label: viewOptionConnection, scope: port.ScopeConnections},
		{key: "A", label: viewOptionAll, scope: port.ScopeAll},
	} {
		marker := markerInactive
		if option.scope == m.Scope {
			marker = markerActive
		}
		fmt.Fprintf(&b, viewMenuRowFmt, marker, option.key, option.label)
	}
	b.WriteString(viewMenuActions)
	return b.String()
}

func (m Model) processName(record model.PortInfo) string {
	if info, ok := m.Infos[record.PID]; ok && strings.TrimSpace(info.Name) != "" {
		return info.Name
	}
	if _, failed := m.LookupErrors[record.PID]; failed {
		return placeholderUnknown
	}
	if strings.TrimSpace(record.ProcessName) != "" {
		return record.ProcessName
	}
	return placeholderUnknown
}

// parentChain renders the ancestor line for the details page using the same
// shared walk as the info command; empty when no parent is known.
func (m Model) parentChain(info model.ProcessInfo) string {
	if m.Manager == nil {
		return ""
	}
	ctx := m.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ancestors := processinfo.Ancestors(ctx, m.Manager, info, processinfo.MaxAncestorHops)
	chain := make([]model.ProcessAncestor, 0, len(ancestors))
	for _, ancestor := range ancestors {
		chain = append(chain, model.ProcessAncestor{PID: ancestor.PID, Name: ancestor.Name})
	}
	return processinfo.FormatAncestors(info.Name, info.PID, chain)
}

func (m Model) visibleIndexes() []int {
	indexes := make([]int, 0, len(m.Ports))
	filter := strings.ToLower(strings.TrimSpace(m.Filter))
	for index, record := range m.Ports {
		if m.PortFilter > 0 && record.Port != m.PortFilter {
			continue
		}
		name := strings.ToLower(m.processName(record))
		if filter != "" &&
			!strings.Contains(name, filter) &&
			!strings.Contains(strconv.Itoa(record.Port), filter) &&
			!strings.Contains(strconv.Itoa(record.PID), filter) {
			continue
		}
		indexes = append(indexes, index)
	}
	return indexes
}

func (m Model) currentKey() rowKey {
	if m.Selected >= 0 && m.Selected < len(m.Ports) {
		return keyOf(m.Ports[m.Selected])
	}
	return m.SelectedKey
}

func sortPorts(ports []model.PortInfo) {
	sort.SliceStable(ports, func(i, j int) bool {
		if ports[i].Port != ports[j].Port {
			return ports[i].Port < ports[j].Port
		}
		if ports[i].Protocol != ports[j].Protocol {
			return ports[i].Protocol < ports[j].Protocol
		}
		if ports[i].PID != ports[j].PID {
			return ports[i].PID < ports[j].PID
		}
		if ports[i].LocalAddr != ports[j].LocalAddr {
			return ports[i].LocalAddr < ports[j].LocalAddr
		}
		return ports[i].RemoteAddr < ports[j].RemoteAddr
	})
}

func restoreSelection(ports []model.PortInfo, previous rowKey, previousIndex int) int {
	for index, record := range ports {
		if keyOf(record) == previous && previous != (rowKey{}) {
			return index
		}
	}
	if len(ports) == 0 {
		return 0
	}
	if previousIndex < 0 {
		return 0
	}
	if previousIndex >= len(ports) {
		return len(ports) - 1
	}
	return previousIndex
}

func formatAge(updated time.Time) string {
	seconds := int(time.Since(updated).Seconds())
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf(ageSecondsFmt, seconds)
}

func displayFilter(filter string) string {
	if strings.TrimSpace(filter) == "" {
		return placeholderDash
	}
	return filter
}

func maxDetailWidth(width int) int {
	if width <= 0 {
		return 58
	}
	available := width - 22
	if available < 16 {
		return 16
	}
	return available
}

func cleanError(err error) string {
	if err == nil {
		return ""
	}
	return strings.Join(strings.Fields(err.Error()), " ")
}

func fitText(value string, width int) string {
	runes := []rune(value)
	if width <= 0 || len(runes) <= width {
		return value
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + ellipsis
}

func display(value string) string {
	if strings.TrimSpace(value) == "" {
		return placeholderUnknown
	}
	return value
}

// Run starts the TUI with the caller-supplied version shown in the header.
func Run(ctx context.Context, scanner port.Scanner, manager process.Manager, version string) error {
	return RunPort(ctx, scanner, manager, 0, version)
}

// RunPort starts the TUI, optionally focused on one listening port. The
// version is injected by the caller so the header always matches the CLI
// --version output.
func RunPort(ctx context.Context, scanner port.Scanner, manager process.Manager, portFilter int, version string) error {
	return runProgramWithPort(ctx, scanner, manager, portFilter, version, tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))
}

func runProgram(ctx context.Context, scanner port.Scanner, manager process.Manager, options ...tea.ProgramOption) error {
	return runProgramWithPort(ctx, scanner, manager, 0, "", options...)
}

func runProgramWithPort(ctx context.Context, scanner port.Scanner, manager process.Manager, portFilter int, version string, options ...tea.ProgramOption) error {
	model := NewWithPort(scanner, manager, portFilter)
	model.Context = ctx
	model.Version = version
	programOptions := append([]tea.ProgramOption{tea.WithContext(ctx)}, options...)
	program := tea.NewProgram(model, programOptions...)
	_, err := program.Run()
	return err
}
