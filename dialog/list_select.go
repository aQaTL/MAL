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

// Displays a list created from given slice / array.
// If allowMultipleSelection is set, you can select multiple entries with space.
func ListSelect(gui *gocui.Gui, title string, slice interface{}, allowMultipleSelection bool) (
	<-chan []int, CleanUpFunc, error,
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
		if bLen := len(strBytes) + idxLen + 2; bLen > listW {
			listW = bLen
		}
		buf.Write(strBytes)
		buf.WriteRune('\n')
	}
	listW += 2

	cleanUp := cleanUpFunc(gui, listSelectViewName)
	selectedIdxes, v, err := listSelect(gui, title, listW, listH, allowMultipleSelection)
	if v != nil {
		buf.WriteTo(v)
	}
	return selectedIdxes, cleanUp, err
}

func listSelect(gui *gocui.Gui, title string, listW, listH int, multiSel bool) (
	chan []int, *gocui.View, error,
) {
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

	selectedIdxes := make(chan []int)
	chanClosed := false

	idxs := make([]int, 0, listH)
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
		case key == gocui.KeySpace:
			if !multiSel {
				return
			}
			_, oy := v.Origin()
			_, y := v.Cursor()
			y += oy
			idxsIdx := -1
			for i, v := range idxs {
				if v == y {
					idxsIdx = i
					break
				}
			}
			v.SetCursor(0, y-oy)
			if idxsIdx != -1 {
				idxs = append(idxs[:idxsIdx], idxs[idxsIdx+1:]...)
				v.EditDelete(false)
			} else {
				idxs = append(idxs, y)
				v.EditWrite('*')
			}
		case key == gocui.KeyEnter:
			if !chanClosed {
				_, oy := v.Origin()
				_, y := v.Cursor()
				y += oy
				contains := false
				for _, v := range idxs {
					if v == y {
						contains = true
						break
					}
				}
				if !contains {
					idxs = append(idxs, y)
				}

				selectedIdxes <- idxs
				close(selectedIdxes)
				chanClosed = true
			}
		case key == gocui.KeyCtrlQ || key == gocui.KeyEsc || ch == 'q':
			if !chanClosed {
				close(selectedIdxes)
				chanClosed = true
			}
		}
	})

	return selectedIdxes, v, err
}
