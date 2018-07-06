package dialog

import "github.com/jroimartin/gocui"

func cleanUpFunc(gui *gocui.Gui, viewToDelete string) func(gui *gocui.Gui) error {
	currView := gui.CurrentView()
	return func(gui *gocui.Gui) error {
		gui.DeleteView(viewToDelete)
		if currView != nil {
			gui.SetCurrentView(currView.Name())
		}
		return nil
	}
}

func longestStringLen(slice []string) (maxLen int) {
	for _, str := range slice {
		if l := len(str); l > maxLen {
			maxLen = l
		}
	}
	return
}