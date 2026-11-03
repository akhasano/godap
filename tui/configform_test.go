package tui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestConfigFormTabCyclesThroughFieldsAndButtons(t *testing.T) {
	for _, mechanism := range authMechanisms {
		t.Run(mechanism.ID, func(t *testing.T) {
			oldApp, oldPanel, oldMechanism := app, appPanel, AuthMechanism
			t.Cleanup(func() { app, appPanel, AuthMechanism = oldApp, oldPanel, oldMechanism })
			app = tview.NewApplication().SetInputCapture(appKeyHandler)
			previous := tview.NewInputField()
			appPanel = tview.NewFlex().AddItem(previous, 0, 1, true)
			app.SetRoot(appPanel, true).SetFocus(previous)
			AuthMechanism = mechanism.ID
			appPanel.SetInputCapture(appPanelKeyHandler)
			screen := tcell.NewSimulationScreen("UTF-8")
			app.SetScreen(screen)
			screen.SetSize(120, 40)
			drawn := make(chan struct{}, 16)
			app.SetAfterDrawFunc(func(tcell.Screen) {
				drawn <- struct{}{}
			})
			done := make(chan error, 1)
			go func() { done <- app.Run() }()
			t.Cleanup(func() {
				app.Stop()
				if err := <-done; err != nil {
					t.Error(err)
				}
			})
			wantFocus := func(want string) {
				t.Helper()
				select {
				case <-drawn:
					got := "unknown"
					switch focus := app.GetFocus().(type) {
					case tview.FormItem:
						got = focus.GetLabel()
					case *tview.Button:
						got = focus.GetLabel()
					}
					if got != want {
						t.Fatalf("focus = %q, want %q", got, want)
					}
				case <-time.After(3 * time.Second):
					t.Fatalf("timed out waiting for focus %q", want)
				}
			}
			wantFocus("")
			app.QueueEvent(tcell.NewEventKey(tcell.KeyF2, 0, tcell.ModNone))
			wantFocus("Server")

			labels := []string{"Server", "Port", "LDAPS", "IgnoreCert", "SOCKSProxy", "Domain", "Auth Type"}
			for _, field := range mechanism.Fields {
				labels = append(labels, string(field))
			}
			labels = append(labels, "Go Back", "Update")
			for i := 1; i <= len(labels); i++ {
				app.QueueEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
				wantFocus(labels[i%len(labels)])
			}
			for i := len(labels) - 1; i >= 0; i-- {
				app.QueueEvent(tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModShift))
				wantFocus(labels[i])
			}

			for i := 1; i < 7; i++ {
				app.QueueEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
				wantFocus(labels[i])
			}
			dropdown := app.GetFocus().(*tview.DropDown)
			app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
			wantFocus("unknown") // The open dropdown delegates focus to its list.
			app.QueueEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
			wantFocus(string(mechanism.Fields[0]))
			if dropdown.IsOpen() {
				t.Fatal("Tab left the auth dropdown open after moving focus")
			}

			// Changing the auth type must replace the credential tab stops.
			app.QueueEvent(tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModShift))
			wantFocus("Auth Type")
			app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
			wantFocus("unknown")
			app.QueueEvent(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
			wantFocus("unknown")
			app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
			wantFocus("Auth Type")
			next := authMechanisms[(indexOfMechanism(mechanism.ID)+1)%len(authMechanisms)]
			for _, field := range next.Fields {
				app.QueueEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
				wantFocus(string(field))
			}
			app.QueueEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
			wantFocus("Go Back")
			app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
			wantFocus("")
		})
	}
}
