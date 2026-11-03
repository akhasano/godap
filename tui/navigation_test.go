package tui

import (
	"testing"

	"github.com/Macmod/godap/v2/pkg/ldaputils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestPageShortcutWrapsAndFocuses(t *testing.T) {
	oldApp, oldPanel, oldInfo, oldPages, oldPage := app, appPanel, info, pages, page
	defer func() { app, appPanel, info, pages, page = oldApp, oldPanel, oldInfo, oldPages, oldPage }()
	app = tview.NewApplication()
	info = tview.NewTextView().SetRegions(true)
	pages = tview.NewPages()
	first, last := tview.NewTable(), tview.NewTable()
	pages.AddPage("page-0", first, true, true).AddPage("page-1", last, true, false)
	appPanel = tview.NewFlex().AddItem(pages, 0, 1, true)
	info.SetHighlightedFunc(func(added, _, _ []string) {
		if len(added) != 0 {
			if added[0] == "0" {
				page = 0
				app.SetFocus(first)
			} else {
				page = 1
				app.SetFocus(last)
			}
			pages.SwitchToPage("page-" + added[0])
		}
	})
	info.SetText(`["0"]First[""] ["1"]Last[""]`)
	app.SetRoot(appPanel, true).SetFocus(first)
	page = 0
	if got := appKeyHandler(tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModNone)); got != nil || page != 1 || app.GetFocus() != last {
		t.Fatalf("Shift+H: event=%v, page=%d, focus=%v", got, page, app.GetFocus())
	}
	if got := appKeyHandler(tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone)); got != nil || page != 0 || app.GetFocus() != first {
		t.Fatalf("Shift+L: event=%v, page=%d, focus=%v", got, page, app.GetFocus())
	}
}

func TestVimNavigationDoesNotStealInputOrModalKeys(t *testing.T) {
	oldApp, oldPanel := app, appPanel
	defer func() { app, appPanel = oldApp, oldPanel }()
	app = tview.NewApplication()
	input, table := tview.NewInputField(), tview.NewTable()
	appPanel = tview.NewFlex().AddItem(input, 0, 1, false).AddItem(table, 0, 1, false)
	app.SetRoot(appPanel, true).SetFocus(input)
	for _, r := range "hjklHL" {
		event := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
		if got := appKeyHandler(event); got != event {
			t.Fatalf("input %q intercepted", r)
		}
	}
	modal := tview.NewInputField()
	app.SetRoot(modal, true).SetFocus(modal)
	event := tcell.NewEventKey(tcell.KeyCtrlL, 0, tcell.ModNone)
	if got := appKeyHandler(event); got != event {
		t.Fatal("modal shortcut intercepted")
	}
}

func TestDirectionalPaneNavigation(t *testing.T) {
	oldApp, oldPanel, oldPage := app, appPanel, page
	oldTree, oldAttrs := treePanel, explorerAttrsPanel
	defer func() {
		app, appPanel, page, treePanel, explorerAttrsPanel = oldApp, oldPanel, oldPage, oldTree, oldAttrs
	}()
	app = tview.NewApplication()
	treePanel, explorerAttrsPanel = tview.NewTreeView(), tview.NewTable()
	appPanel = tview.NewFlex().AddItem(treePanel, 0, 1, false).AddItem(explorerAttrsPanel, 0, 1, false)
	app.SetRoot(appPanel, true).SetFocus(treePanel)
	page = 0
	if got := appKeyHandler(tcell.NewEventKey(tcell.KeyCtrlL, 0, tcell.ModNone)); got != nil || app.GetFocus() != explorerAttrsPanel {
		t.Fatal("Ctrl+L should focus attributes")
	}
	if got := appKeyHandler(tcell.NewEventKey(tcell.KeyCtrlH, 0, tcell.ModNone)); got != nil || app.GetFocus() != treePanel {
		t.Fatal("Ctrl+H should focus tree")
	}
}

func TestTreeHAndLActLikeArrows(t *testing.T) {
	oldApp, oldPanel, oldTree := app, appPanel, treePanel
	defer func() { app, appPanel, treePanel = oldApp, oldPanel, oldTree }()
	app = tview.NewApplication()
	treePanel = tview.NewTreeView()
	appPanel = tview.NewFlex().AddItem(treePanel, 0, 1, false)
	app.SetRoot(appPanel, true).SetFocus(treePanel)
	for r, want := range map[rune]tcell.Key{'h': tcell.KeyLeft, 'l': tcell.KeyRight} {
		got := appKeyHandler(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
		if got == nil || got.Key() != want {
			t.Fatalf("%q: got %v, want %v", r, got, want)
		}
	}
}

func TestSearchPaneNavigationUsesVisibleSidePage(t *testing.T) {
	oldPage, oldQuery, oldTree, oldSide := page, searchQueryPanel, searchTreePanel, sidePanel
	defer func() { page, searchQueryPanel, searchTreePanel, sidePanel = oldPage, oldQuery, oldTree, oldSide }()
	page = 1
	searchQueryPanel, searchTreePanel = tview.NewInputField(), tview.NewTreeView()
	sidePanel = tview.NewPages()
	hidden, visible := tview.NewTable(), tview.NewTable()
	sidePanel.AddPage("hidden", hidden, true, false).AddPage("visible", visible, true, true)
	if got := directionalPane(searchTreePanel, tcell.KeyCtrlL); got != visible {
		t.Fatalf("Ctrl+L targeted %v instead of visible side panel", got)
	}
	if got := directionalPane(visible, tcell.KeyCtrlH); got != searchTreePanel {
		t.Fatal("Ctrl+H should return to search results")
	}
	if got := directionalPane(searchQueryPanel, tcell.KeyCtrlJ); got != searchTreePanel {
		t.Fatal("Ctrl+J should go down to search results")
	}
	if got := directionalPane(searchTreePanel, tcell.KeyCtrlK); got != searchQueryPanel {
		t.Fatal("Ctrl+K should go up to search filter")
	}
}

func TestGroupTabIncludesDepthInput(t *testing.T) {
	oldApp, oldDepth, oldGroup, oldObject, oldMembers, oldGroups := app, depthInput, groupNameInput, objectNameInput, membersPanel, groupsPanel
	defer func() {
		app, depthInput, groupNameInput, objectNameInput, membersPanel, groupsPanel = oldApp, oldDepth, oldGroup, oldObject, oldMembers, oldGroups
	}()
	oldFlavor := lc.Flavor
	defer func() { lc.Flavor = oldFlavor }()
	lc.Flavor = ldaputils.MicrosoftADFlavor
	app = tview.NewApplication()
	depthInput, groupNameInput, objectNameInput = tview.NewInputField(), tview.NewInputField(), tview.NewInputField()
	membersPanel, groupsPanel = tview.NewTable(), tview.NewTable()
	app.SetFocus(depthInput)
	groupRotateFocus()
	if app.GetFocus() != groupNameInput {
		t.Fatal("Tab should go from depth to group")
	}
	app.SetFocus(groupsPanel)
	groupRotateFocus()
	if app.GetFocus() != depthInput {
		t.Fatal("Tab should wrap to depth")
	}
}

func TestDNSTabFocusesVisibleSidePane(t *testing.T) {
	oldApp, oldFilter, oldSide, oldProps, oldRecords, oldTree := app, dnsZoneFilter, dnsSidePanel, dnsZoneProps, dnsNodeRecords, dnsTreePanel
	defer func() {
		app, dnsZoneFilter, dnsSidePanel, dnsZoneProps, dnsNodeRecords, dnsTreePanel = oldApp, oldFilter, oldSide, oldProps, oldRecords, oldTree
	}()
	app = tview.NewApplication()
	dnsZoneFilter, dnsZoneProps, dnsNodeRecords, dnsTreePanel = tview.NewInputField(), tview.NewTable(), tview.NewTreeView(), tview.NewTreeView()
	dnsSidePanel = tview.NewPages().AddPage("props", dnsZoneProps, true, false).AddPage("records", dnsNodeRecords, true, true)
	app.SetFocus(dnsZoneFilter)
	dnsRotateFocus()
	if app.GetFocus() != dnsNodeRecords {
		t.Fatal("Tab should focus visible records pane")
	}
	dnsRotateFocus()
	if app.GetFocus() != dnsTreePanel {
		t.Fatal("Tab should return to zones tree")
	}
}
