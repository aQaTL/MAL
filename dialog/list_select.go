package dialog

import (
	"bytes"
	"fmt"
	"github.com/jroimartin/gocui"
	"strconv"
)

const (
	listSelectViewName = "listSelectViewName"
)

func ListSelect(gui *gocui.Gui, title string, list []fmt.Stringer) (
	<-chan int, func(gui2 *gocui.Gui) error, error,
) {
	listW := 1
	listH := len(list)

	buffer := bytes.Buffer{}

	for i, str := range list {
		buffer.WriteString(strconv.Itoa(i))
		buffer.Write([]byte{'.', ' '})
		strBytes := []byte(str.String())
		if bLen := len(strBytes); bLen > listW {
			listW = bLen
		}
		buffer.Write(strBytes)
		buffer.WriteRune('\n')
	}
	listW += 2

	cleanUp := cleanUpFunc(gui)
	selectedIdx, v, err := listSelect(gui, title, listW, listH)
	if v != nil {
		buffer.WriteTo(v)
	}
	return selectedIdx, cleanUp, err
}

func ListSelectString(gui *gocui.Gui, title string, list []string) (
	<-chan int, func(*gocui.Gui) error, error,
) {
	listW := longestStringLen(list) + 2
	listH := len(list)

	cleanUp := cleanUpFunc(gui)
	selectedIdx, v, err := listSelect(gui, title, listW, listH)

	if v != nil {
		for i, str := range list {
			fmt.Fprint(v, i, ".")
			fmt.Fprintln(v, "", str)
		}
	}

	return selectedIdx, cleanUp, err
}

func listSelect(gui *gocui.Gui, title string, listW, listH int) (chan int, *gocui.View, error) {
	w, h := gui.Size()

	//TODO scrolling list if list length is too big
	//TODO wrap list if list is too wide (handle moving up & down correctly)
	x0, y0 := w/2-listW/2, h/2-(listH+1)/2
	x1, y1 := x0+listW, y0+(listH+1)

	v, err := gui.SetView(listSelectViewName, x0, y0, x1, y1)
	if err == gocui.ErrUnknownView {
		err = nil
	}

	gui.SetCurrentView(listSelectViewName)
	gui.SetViewOnTop(listSelectViewName)

	v.Title = title
	v.SelBgColor = gocui.ColorGreen
	v.SelBgColor = gocui.ColorGreen
	v.SelFgColor = gocui.ColorBlack
	v.Highlight = true
	v.Editable = true

	selectedIdx := make(chan int)
	chanClosed := false

	v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
		switch {
		case key == gocui.KeyArrowDown || ch == 'j':
			_, oy := v.Origin()
			_, y := v.Cursor()
			y += oy
			if y < listH-1 {
				v.MoveCursor(0, 1, false)
			}
		case key == gocui.KeyArrowUp || ch == 'k':
			v.MoveCursor(0, -1, false)
		case key == gocui.KeyEnter:
			//TODO mutliple selection (color / highlight selected row) (buffered channels?)
			_, oy := v.Origin()
			_, y := v.Cursor()
			y += oy
			if !chanClosed {
				selectedIdx <- y
				close(selectedIdx)
				chanClosed = true
			}
		case key == gocui.KeyCtrlQ || key == gocui.KeyEsc || ch == 'q':
			if !chanClosed {
				close(selectedIdx)
				chanClosed = true
			}
		}
	})

	return selectedIdx, v, err
}

func cleanUpFunc(gui *gocui.Gui) func(gui *gocui.Gui) error {
	currView := gui.CurrentView()
	return func(gui *gocui.Gui) error {
		gui.DeleteView(listSelectViewName)
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