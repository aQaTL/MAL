package main

import (
	"fmt"
	"github.com/aqatl/mal/nyaa_scraper"
	"github.com/jroimartin/gocui"
	"github.com/urfave/cli"
	"os/exec"
	"time"
	"unicode/utf8"
)

func browseNyaa(ctx *cli.Context) error {
	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	entry := list.GetByID(cfg.SelectedID)
	if entry == nil {
		return fmt.Errorf("no entry found")
	}

	gui, err := gocui.NewGui(gocui.OutputNormal)
	defer gui.Close()
	if err != nil {
		return fmt.Errorf("gocui error: %v", err)
	}

	pl := &nyaaPageLoaderCui{
		SearchTerm: entry.Title,
		Category:   nyaa_scraper.AnimeEnglishTranslated,
		Filter:     nyaa_scraper.NoFilter,
		PageToLoad: 1,
	}
	done := setUpPageLoaderCui(gui, pl)

	go func() {
		<-done

		gui.Update(func(gui *gocui.Gui) error {
			if pl.ResultErr != nil {
				return pl.ResultErr
			}

			nc := &nyaaCui{
				Cfg:         cfg,
				ResultPages: []nyaa_scraper.NyaaResultPage{pl.Result},
				CurrPageIdx: pl.PageToLoad - 1,
			}
			setUpNyaaCui(nc, gui)

			return nil
		})
	}()

	if err = gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}

	return nil
}

const (
	ncInfoView      = "ncInfoView"
	ncResultsView   = "ncResultsView "
	ncShortcutsView = "ncShortcutsView"
)

type nyaaCui struct {
	Cfg *Config

	ResultPages []nyaa_scraper.NyaaResultPage
	CurrPageIdx int

	ResultsView *gocui.View
}

func setUpNyaaCui(nc *nyaaCui, gui *gocui.Gui) {
	gui.SetManager(nc)
	nc.setGuiKeyBindings(gui)

	gui.Cursor = false
	gui.Mouse = false
	gui.Highlight = true
	gui.SelFgColor = gocui.ColorGreen
}

func (nc *nyaaCui) Layout(gui *gocui.Gui) error {
	w, h := gui.Size()
	if v, err := gui.SetView(ncInfoView, 0, 0, w-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Info"
		v.Editable = false

		currPage := &nc.ResultPages[nc.CurrPageIdx]
		fmt.Fprintf(v, "Displaying results %d-%d out of %d results",
			currPage.DisplayedFrom, currPage.DisplayedTo, currPage.DisplayedOutOf)
	}

	if v, err := gui.SetView(ncResultsView, 0, 3, w-1, h-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Search results"
		v.SelBgColor = gocui.ColorGreen
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = true
		v.Editable = true
		v.Editor = gocui.EditorFunc(nc.Editor(gui))

		gui.SetCurrentView(ncResultsView)
		nc.ResultsView = v

		for i, result := range nc.ResultPages[nc.CurrPageIdx].Results {
			fmt.Fprintf(v, "%d. %s %s %v %d %d %d\n",
				i+1,
				result.Title,
				result.Size,
				result.DateAdded,
				result.Seeders,
				result.Leechers,
				result.CompletedDownloads,
			)
		}
	}

	if v, err := gui.SetView(ncShortcutsView, 0, h-3, w-1, h-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Shortcuts"
		v.Editable = false

		fmt.Fprintf(v, "d download")
	}

	return nil
}

func (nc *nyaaCui) Editor(gui *gocui.Gui) func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	return func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
		switch {
		case key == gocui.KeyArrowDown || ch == 'j':
			v.MoveCursor(0, 1, false)
		case key == gocui.KeyArrowUp || ch == 'k':
			v.MoveCursor(0, -1, false)
		case ch == 'd':
			_, y := v.Cursor()
			_, oy := v.Origin()
			y += oy
			if ml := len(nc.ResultPages[nc.CurrPageIdx].Results); ml == 0 || y > ml-1 || y < 0 {
				return
			}

			link := ""
			if entry := nc.ResultPages[nc.CurrPageIdx].Results[y]; entry.MagnetLink != "" {
				link = entry.MagnetLink
			} else if entry.TorrentLink != "" {
				link = entry.TorrentLink
			} else {
				gui.Update(func(gui *gocui.Gui) error {
					//TODO don't exit app when no link is present
					return fmt.Errorf("no link found")
				})
				return
			}

			if err := nc.Download(link); err != nil {
				gui.Update(func(gui *gocui.Gui) error {
					return err
				})
			}
		}
	}
}

func (nc *nyaaCui) Download(link string) error {
	link = "\"" + link + "\""
	cmd := exec.Command(nc.Cfg.TorrentClientPath, nc.Cfg.TorrentClientArgs, link)
	cmd.Args = cmd.Args[1:]
	return cmd.Start()
}

func (nc *nyaaCui) setGuiKeyBindings(gui *gocui.Gui) {
	gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitGocui)
}

const (
	nplcStatusView = "nplcStatusView"
)

type nyaaPageLoaderCui struct {
	SearchTerm string
	Category   nyaa_scraper.NyaaCategory
	Filter     nyaa_scraper.NyaaFilter
	PageToLoad int

	Result    nyaa_scraper.NyaaResultPage
	ResultErr error

	doneInner chan struct{}
}

func setUpPageLoaderCui(gui *gocui.Gui, pl *nyaaPageLoaderCui) (done chan struct{}) {
	pl.doneInner = make(chan struct{})
	done = make(chan struct{})
	go func(pl *nyaaPageLoaderCui) {
		result, err := nyaa_scraper.SearchSpecificPage(
			pl.SearchTerm, pl.Category, pl.Filter, pl.PageToLoad,
		)
		pl.Result = result
		pl.ResultErr = err

		pl.doneInner <- struct{}{}
		close(pl.doneInner)
		done <- struct{}{}
		close(done)
	}(pl)

	gui.SetManager(pl)
	gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitGocui)

	gui.Cursor = false
	gui.Highlight = true
	gui.Mouse = false

	return
}

func (pl *nyaaPageLoaderCui) Layout(gui *gocui.Gui) error {
	w, h := gui.Size()
	vw, vh := 19+utf8.RuneCountInString(pl.SearchTerm), 2
	x0, y0 := w/2-vw/2, h/2-vh/2
	x1, y1 := x0+vw, y0+vh
	if v, err := gui.SetView(nplcStatusView, x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Highlight = true
		v.Editable = false

		fmt.Fprintf(v, "Searching for %s [-]", pl.SearchTerm)

		gui.SetCurrentView(nplcStatusView)

		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			clockStates := [...]string{"-", "\\", "|", "/"}
			currClockState := 1

		loop:
			for {
				select {
				case <-ticker.C:
					gui.Update(func(gui *gocui.Gui) error {
						v.Clear()
						fmt.Fprintf(v, "Searching for %s [%s]",
							pl.SearchTerm,
							clockStates[currClockState])

						currClockState = (currClockState + 1) % len(clockStates)
						return nil
					})
				case <-pl.doneInner:
					break loop
				}
			}
		}()
	}

	return nil
}

func quitGocui(gui *gocui.Gui, view *gocui.View) error {
	return gocui.ErrQuit
}
