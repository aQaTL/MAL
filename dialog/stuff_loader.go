package dialog

import (
	"github.com/jroimartin/gocui"
	"fmt"
	"unicode/utf8"
)

const stuffLoaderViewName = "stuffLoaderViewName"

func StuffLoader(config Config, f func()) (<-chan bool, error) {
	cleanUp := cleanUpFunc(config.Gui, stuffLoaderViewName)
	v, err := config.Gui.SetView(stuffLoaderViewName, config.X0, config.Y0, config.X1, config.Y1)
	if err == gocui.ErrUnknownView {
		err = nil
	} else if err != nil {
		return nil, err
	}

	jobDone := make(chan bool)

	v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
		switch {
		case key == gocui.KeyCtrlQ || key == gocui.KeyEsc || ch == 'q':
			jobDone <- false
			config.Gui.Update(cleanUp)
		}
	})

	config.Gui.SetCurrentView(stuffLoaderViewName)
	config.Gui.SetViewOnTop(stuffLoaderViewName)
	defaultViewConfig(v)
	config.ViewConfig(v)

	go func() {
		f()
		jobDone <- true
		config.Gui.Update(cleanUp)
	}()

	return jobDone, nil
}

func FitMessage(gui *gocui.Gui, msg string, cfgs ...func(*gocui.View)) Config {
	w, h := gui.Size()
	vw, vh := utf8.RuneCountInString(msg)+1, 2
	x0, y0 := w/2-vw/2, h/2-vh/2
	x1, y1 := x0+vw, y0+vh
	return Config{
		gui,
		Pos{x0, y0, x1, y1},
		func(v *gocui.View) {
			for _, cfg := range cfgs {
				cfg(v)
			}
			Msg(msg)(v)
		}}
}

func defaultViewConfig(v *gocui.View) {
	v.Editable = true
	v.Highlight = true
	v.Wrap = true
}

func Msg(msg string) func(*gocui.View) {
	return func(v *gocui.View) {
		fmt.Fprintln(v, msg)
	}
}

func Title(title string) func(*gocui.View) {
	return func(v *gocui.View) {
		v.Title = title
	}
}
