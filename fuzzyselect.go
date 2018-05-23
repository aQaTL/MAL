package main

import (
	"bufio"
	"fmt"
	"github.com/aqatl/mal/mal"
	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/sahilm/fuzzy"
	"github.com/urfave/cli"
	"strings"
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
		listStr[i] = strings.ToLower(fmt.Sprintf("%s %s",
			entry.Title,
			strings.Replace(entry.Synonyms, ";", "", -1)))
	}

	fsg := &fuzzySelGui{
		List:    list,
		ListStr: listStr,
		Cfg:     cfg,
		Match:   nil,
	}

	gui.SetManager(fsg)
	setGuiKeyBindings(gui, fsg)

	gui.Cursor = true
	gui.Mouse = false
	gui.Highlight = true
	gui.SelFgColor = gocui.ColorGreen

	initSearch := strings.Join(ctx.Args(), " ")
	fsg.Layout(gui)
	fsg.InputView.Write([]byte(initSearch))
	fsg.InputView.Editor.Edit(fsg.InputView, gocui.KeyArrowUp, 0, gocui.ModNone)

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
	FsgInputView     = "fsgInputView"
	FsgOutputView    = "fsgOutputView"
	FsgShortcutsView = "fsgShortcutsView"
)

type fuzzySelGui struct {
	List    mal.AnimeList
	ListStr []string
	Cfg     *Config

	Matches []fuzzy.Match
	Match   *mal.Anime

	InputView, OutputView *gocui.View
}

func (fsg *fuzzySelGui) Layout(gui *gocui.Gui) error {
	w, h := gui.Size()
	if v, err := gui.SetView(FsgInputView, 0, 0, w-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Input"
		v.Editor = gocui.EditorFunc(fsg.InputViewEditor)
		v.Editable = true
		v.Wrap = true

		gui.SetCurrentView(FsgInputView)
		fsg.InputView = v
	}

	if v, err := gui.SetView(FsgOutputView, 0, 3, w-1, h-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Found entries"
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = true
		v.Editable = true
		v.Editor = gocui.EditorFunc(fsg.OutputViewEditor)

		fsg.OutputView = v
	}

	if v, err := gui.SetView(FsgShortcutsView, 0, h-3, w-1, h-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Shortcuts"
		v.Editable = false

		fmt.Fprintf(v, "Ctrl+C: quit | Tab: switch window | Enter: select highlighted entry")
	}

	return nil
}

func (fsg *fuzzySelGui) InputViewEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	gocui.DefaultEditor.Edit(v, key, ch, mod)

	fsg.OutputView.Clear()

	pattern := strings.TrimSpace(v.Buffer())
	fsg.Matches = fuzzy.Find(pattern, fsg.ListStr)

	highlight := color.New(color.FgBlack, color.BgWhite).FprintFunc()
	buf := bufio.NewWriter(fsg.OutputView)

	for _, match := range fsg.Matches {
		mIdx := 0
		for i, r := range []rune(fsg.List[match.Index].Title) {
			if mIdx < len(match.MatchedIndexes) && i == match.MatchedIndexes[mIdx] {
				mIdx++
				highlight(buf, string(r))
			} else {
				buf.WriteRune(r)
			}
		}
		buf.WriteRune('\n')
	}
	buf.Flush()

	fsg.OutputView.SetCursor(0, 0)
}

func (fsg *fuzzySelGui) OutputViewEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case key == gocui.KeyArrowDown || ch == 'j':
		v.MoveCursor(0, 1, false)
	case key == gocui.KeyArrowUp || ch == 'k':
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

	gui.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, v *gocui.View) error {
		_, y := fsg.OutputView.Cursor()
		if ml := len(fsg.Matches); ml == 0 || y > ml-1 || y < 0 {
			return nil
		}

		fsg.Match = fsg.List[fsg.Matches[y].Index]

		return gocui.ErrQuit
	})
}
