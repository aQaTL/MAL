package dialog

import "github.com/jroimartin/gocui"

/*
(x0,y0)---------(  ,  )
   |               |
   |               |
   |               |
(  ,  )---------(x1,y1)
 */
type Pos struct {
	X0, Y0 int
	X1, Y1 int
}

type Config struct {
	Gui        *gocui.Gui
	Pos
	ViewConfig func(*gocui.View)
}

type CleanUpFunc func(gui *gocui.Gui) error

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