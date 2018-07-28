package dialog

import (
	"bytes"
	"fmt"
	"github.com/jroimartin/gocui"
	"reflect"
	"strconv"
)

const (
	listSelectViewName = "listSelectViewName"
)

func ListSelect(gui *gocui.Gui, title string, slice interface{}) (
	<-chan int, CleanUpFunc, error,
) {
	value := reflect.ValueOf(slice)
	if k := value.Kind(); k != reflect.Slice && k != reflect.Array {
		return nil, nil, fmt.Errorf("slice argument must be a slice or an array")
	}
	listW := 1
	listH := value.Len()

	buf := bytes.Buffer{}
	for i := 0; i < value.Len(); i++ {
		idxLen, _ := buf.WriteString(strconv.Itoa(i))
		buf.WriteByte('.')
		strBytes := []byte(fmt.Sprint(value.Index(i).Interface()))
		if bLen := len(strBytes) + idxLen + 1; bLen > listW {
			listW = bLen
		}
		buf.Write(strBytes)
		buf.WriteRune('\n')
	}
	listW += 2

	cleanUp := cleanUpFunc(gui, listSelectViewName)
	selectedIdx, v, err := listSelect(gui, title, listW, listH)
	if v != nil {
		buf.WriteTo(v)
	}
	return selectedIdx, cleanUp, err
}

func listSelect(gui *gocui.Gui, title string, listW, listH int) (chan int, *gocui.View, error) {
	w, h := gui.Size()

	//TODO wrap list if list is too wide (handle moving up & down correctly)
	viewHeight := listH + 1
	if maxHeight := int(float64(h) * 0.8); listH > maxHeight {
		viewHeight = maxHeight
	}
	x0, y0 := w/2-listW/2, h/2-viewHeight/2
	x1, y1 := x0+listW, y0+viewHeight

	v, err := gui.SetView(listSelectViewName, x0, y0, x1, y1)
	if err == gocui.ErrUnknownView {
		err = nil
	}

	gui.SetCurrentView(listSelectViewName)
	gui.SetViewOnTop(listSelectViewName)

	v.Title = title
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
