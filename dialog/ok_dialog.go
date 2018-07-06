package dialog

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"math"
	"strings"
	"unicode/utf8"
)

const okDialogViewName = "okDialogViewName "

func JustShowOkDialog(gui *gocui.Gui, title, msg string) {
	gui.Update(func(gui *gocui.Gui) error {
		confirmed, cleanUp, err := OkDialog(FitMessageWithOkButton(
			gui,
			msg,
			Title(title),
		))
		if err != nil {
			return err
		}
		go func() {
			<-confirmed
			gui.Update(cleanUp)
		}()
		return nil
	})
}

func OkDialog(config Config) (<-chan struct{}, CleanUpFunc, error) {
	cleanUp := cleanUpFunc(config.Gui, okDialogViewName)
	v, err := config.Gui.SetView(okDialogViewName, config.X0, config.Y0, config.X1, config.Y1)
	if err == gocui.ErrUnknownView {
		err = nil
	} else if err != nil {
		return nil, cleanUp, err
	}

	confirmed := make(chan struct{})

	v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
		switch key {
		case gocui.KeyEnter, gocui.KeySpace:
			confirmed <- struct{}{}
		}
	})

	config.Gui.SetCurrentView(okDialogViewName)
	config.Gui.SetViewOnTop(okDialogViewName)
	v.Editable = true
	v.Highlight = false
	v.Wrap = true
	config.ViewConfig(v)

	return confirmed, cleanUp, nil
}

func FitMessageWithOkButton(gui *gocui.Gui, msg string, cfgs ...func(*gocui.View)) Config {
	//TODO support long messages
	w, h := gui.Size()
	vw, vh := int(math.Max(float64(utf8.RuneCountInString(msg)+2), 5)), 3
	x0, y0 := w/2-vw/2, h/2-vh/2
	x1, y1 := x0+vw, y0+vh
	return Config{
		gui,
		Pos{x0, y0, x1, y1},
		func(v *gocui.View) {
			for _, cfg := range cfgs {
				cfg(v)
			}

			fmt.Fprintln(v, msg)
			filler := strings.Repeat(" ", vw/2-2)
			fmt.Fprint(v, filler, ">OK<", filler)

			v.SetCursor(0, 1)
			v.Highlight = true
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack
		}}
}
