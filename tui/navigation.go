package tui

import (
	"github.com/Macmod/godap/v2/pkg/ldaputils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// panePosition describes the focusable parts of a page, not decorative views.
// Positions follow the visible layout so directional shortcuts never wrap.
type panePosition struct {
	view tview.Primitive
	row  int
	col  int
}

func visiblePane(p *tview.Pages) tview.Primitive {
	if p == nil {
		return nil
	}
	_, view := p.GetFrontPage()
	return view
}

func currentPanes() []panePosition {
	switch page {
	case 0:
		return []panePosition{{treePanel, 0, 0}, {explorerAttrsPanel, 0, 1}}
	case 1:
		return []panePosition{{searchQueryPanel, 0, 0}, {searchTreePanel, 1, 0}, {visiblePane(sidePanel), 1, 1}}
	case 2:
		panes := []panePosition{{groupNameInput, 1, 0}, {objectNameInput, 1, 1}, {membersPanel, 2, 0}, {groupsPanel, 2, 1}}
		if lc.Flavor == ldaputils.MicrosoftADFlavor { // The depth field exists only in the AD layout.
			panes = append(panes, panePosition{depthInput, 0, 0})
		}
		return panes
	case 3:
		if lc.Flavor == ldaputils.BasicLDAPFlavor {
			return []panePosition{{keybindingsPanel, 0, 0}}
		}
		return []panePosition{{objectNameInputDacl, 0, 0}, {daclEntriesPanel, 1, 0}}
	case 4:
		return []panePosition{{gpoTargetInput, 0, 0}, {gpoListPanel, 1, 0}, {gpoPath, 1, 1}, {gpoLinksPanel, 2, 1}}
	case 5:
		return []panePosition{{dnsQueryPanel, 0, 0}, {dnsNodeFilter, 0, 1}, {dnsZoneFilter, 0, 2}, {dnsTreePanel, 1, 0}, {visiblePane(dnsSidePanel), 1, 1}}
	case 6:
		return []panePosition{{keybindingsPanel, 0, 0}}
	}
	return nil
}

func directionalPane(focus tview.Primitive, key tcell.Key) tview.Primitive {
	panes := currentPanes()
	for _, from := range panes {
		if from.view == nil || from.view != focus {
			continue
		}
		var target tview.Primitive
		best := int(^uint(0) >> 1)
		for _, to := range panes {
			if to.view == nil || to.view == from.view {
				continue
			}
			distance := 0
			switch key {
			case tcell.KeyCtrlH:
				if to.row != from.row || to.col >= from.col {
					continue
				}
				distance = from.col - to.col
			case tcell.KeyCtrlL:
				if to.row != from.row || to.col <= from.col {
					continue
				}
				distance = to.col - from.col
			case tcell.KeyCtrlK:
				if to.col != from.col || to.row >= from.row {
					continue
				}
				distance = from.row - to.row
			case tcell.KeyCtrlJ:
				if to.col != from.col || to.row <= from.row {
					continue
				}
				distance = to.row - from.row
			default:
				return nil
			}
			if distance < best {
				best, target = distance, to.view
			}
		}
		return target
	}
	return nil
}
