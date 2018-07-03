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

	fsc := &fuzzySelCui{
		List:    list,
		ListStr: listStr,
		Cfg:     cfg,
		Match:   nil,
	}

	initSearch := strings.Join(ctx.Args(), " ")

	if ctx.NArg() != 0 {
		fsc.Matches = fuzzy.Find(initSearch, fsc.ListStr)
		if matchesLen := len(fsc.Matches); matchesLen == 0 {
			return fmt.Errorf("no match found")
		} else if matchesLen == 1 {
			saveSelection(cfg, fsc.List[fsc.Matches[0].Index])
			return nil
		}
	}

	if err = startFuzzySelectCUI(fsc, initSearch); err != nil || fsc.Match == nil {
		return err
	}
	saveSelection(cfg, fsc.Match)

	return nil
}

func startFuzzySelectCUI(fsc *fuzzySelCui, initSearch string) error {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return fmt.Errorf("gocui error: %v", err)
	}

	gui.SetManager(fsc)
	fsc.setGuiKeyBindings(gui)

	gui.Cursor = true
	gui.Mouse = false
	gui.Highlight = true
	gui.SelFgColor = gocui.ColorGreen

	fsc.Layout(gui)
	fsc.InputView.Write([]byte(initSearch))
	fsc.InputView.Editor.Edit(fsc.InputView, gocui.KeyBackspace, 0, gocui.ModNone)
	fsc.InputView.MoveCursor(len(initSearch), 0, true)

	err = gui.MainLoop()
	gui.Close()
	if err == gocui.ErrQuit {
		err = nil
	}
	return err
}

func saveSelection(cfg *Config, entry *mal.Anime) {
	cfg.SelectedID = entry.ID
	cfg.Save()

	fmt.Println("Selected entry:")
	printEntryDetails(entry)
}

const (
	FsgInputView     = "fsgInputView"
	FsgOutputView    = "fsgOutputView"
	FsgShortcutsView = "fsgShortcutsView"
)

type fuzzySelCui struct {
	List    mal.AnimeList
	ListStr []string
	Cfg     *Config

	Matches []fuzzy.Match
	Match   *mal.Anime

	InputView, OutputView *gocui.View
}

func (fsc *fuzzySelCui) Layout(gui *gocui.Gui) error {
	w, h := gui.Size()
	if v, err := gui.SetView(FsgInputView, 0, 0, w-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Input"
		v.Editor = gocui.EditorFunc(fsc.InputViewEditor)
		v.Editable = true
		v.Wrap = true

		gui.SetCurrentView(FsgInputView)
		fsc.InputView = v
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
		v.Editor = gocui.EditorFunc(fsc.OutputViewEditor)

		fsc.OutputView = v
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

var highlighter = color.New(color.FgBlack, color.BgWhite).FprintFunc()

func (fsc *fuzzySelCui) InputViewEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if key == gocui.KeyArrowUp || key == gocui.KeyArrowDown {
		fsc.OutputViewEditor(fsc.OutputView, key, ch, mod)
		return
	}
	gocui.DefaultEditor.Edit(v, key, ch, mod)

	fsc.OutputView.Clear()

	pattern := strings.TrimSpace(v.Buffer())
	fsc.Matches = fuzzy.Find(pattern, fsc.ListStr)

	buf := bufio.NewWriter(fsc.OutputView)

	for _, match := range fsc.Matches {
		mIdx := 0
		for i, r := range []rune(fsc.List[match.Index].Title) {
			if mIdx < len(match.MatchedIndexes) && i == match.MatchedIndexes[mIdx] {
				mIdx++
				highlighter(buf, string(r))
			} else {
				buf.WriteRune(r)
			}
		}
		buf.WriteRune('\n')
	}
	buf.Flush()

	fsc.OutputView.SetCursor(0, 0)
}

func (fsc *fuzzySelCui) OutputViewEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case key == gocui.KeyArrowDown || ch == 'j':
		v.MoveCursor(0, 1, false)
	case key == gocui.KeyArrowUp || ch == 'k':
		v.MoveCursor(0, -1, false)
	}
}

func (fsc *fuzzySelCui) setGuiKeyBindings(gui *gocui.Gui) {
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
		_, y := fsc.OutputView.Cursor()
		_, oy := fsc.OutputView.Origin()
		y += oy
		if ml := len(fsc.Matches); ml == 0 || y > ml-1 || y < 0 {
			return nil
		}

		fsc.Match = fsc.List[fsc.Matches[y].Index]

		return gocui.ErrQuit
	})
}
