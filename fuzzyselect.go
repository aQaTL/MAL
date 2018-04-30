package main

import (
	"fmt"
	"github.com/aqatl/mal/mal"
	"github.com/jroimartin/gocui"
	"github.com/urfave/cli"
	"strings"
	"github.com/sahilm/fuzzy"
)

func fuzzySelectEntry(ctx *cli.Context) error {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return fmt.Errorf("gocui error: %v", err)
	}

	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	listStr := make([]string, len(list))
	for i, entry := range list {
		listStr[i] = strings.ToLower(fmt.Sprintf("%s (%s)", entry.Title, entry.Synonyms))
	}

	fsg := &fuzzySelGui{
		List: list,
		ListStr: listStr,
		Cfg:  cfg,
		Match: nil,
	}

	gui.SetManager(fsg)
	setGuiKeyBindings(gui, fsg)

	gui.Cursor = true
	gui.Mouse = false
	gui.Highlight = true
	gui.SelFgColor = gocui.ColorGreen

	err = gui.MainLoop()
	if err != nil && err != gocui.ErrQuit {
		gui.Close()
		return err
	}

	if fsg.Match == nil {
		gui.Close()
		return nil
	}

	cfg.SelectedID = fsg.Match.ID
	cfg.Save()

	gui.Close()

	fmt.Println("Selected entry:")
	printEntryDetails(fsg.Match)

	return nil
}

const (
	FsgInputView  = "fsgInputView"
	FsgOutputView = "fsgOutputView"
)

type fuzzySelGui struct {
	List mal.AnimeList
	ListStr []string
	Cfg  *Config

	Matches []fuzzy.Match
	Match *mal.Anime

	InputView, OutputView *gocui.View
}

func (fsg *fuzzySelGui) Layout(gui *gocui.Gui) error {
	w, h := gui.Size()
	if v, err := gui.SetView(FsgInputView, 0, 0, w-1, h/7); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Input"
		v.Editor = gocui.EditorFunc(fsg.InputViewEditor)
		v.Editable = true
		v.Wrap = true

		fsg.InputView = v
		gui.SetCurrentView(FsgInputView)
	}

	if v, err := gui.SetView(FsgOutputView, 0, h/7+1, w-1, h-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = true

		v.Editable = true
		v.Editor = gocui.EditorFunc(fsg.OutputViewEditor)

		fsg.OutputView = v
	}

	return nil
}

func (fsg *fuzzySelGui) InputViewEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	gocui.DefaultEditor.Edit(v, key, ch, mod)

	fsg.OutputView.Clear()

	pattern := strings.TrimSpace(v.Buffer())

	fsg.Matches = fuzzy.Find(pattern, fsg.ListStr)
	for _, match := range fsg.Matches {
		fmt.Fprintf(fsg.OutputView, "%s\n", fsg.List[match.Index].Title)
	}
}

func (fsg *fuzzySelGui) OutputViewEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch key {
	case gocui.KeyArrowDown:
		v.MoveCursor(0, 1, false)
	case gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	}
}

func setGuiKeyBindings(gui *gocui.Gui, fsg *fuzzySelGui) {
	quit := func(gui *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}
	gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	gui.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, quit)

	gui.SetKeybinding("", gocui.KeyTab, gocui.ModNone, func(gui *gocui.Gui, v *gocui.View) error {
		switch v.Name() {
		case FsgInputView:
			gui.SetCurrentView(FsgOutputView)
		case FsgOutputView:
			gui.SetCurrentView(FsgInputView)
		}
		return nil
	})

	gui.SetKeybinding(FsgOutputView, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, v *gocui.View) error {
		_, y := fsg.OutputView.Cursor()
		if ml := len(fsg.Matches); ml == 0 || y > ml - 1 || y < 0 {
			return nil
		}

		fsg.Match = fsg.List[fsg.Matches[y].Index]

		return gocui.ErrQuit
	})
}
