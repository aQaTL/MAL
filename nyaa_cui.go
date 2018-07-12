package main

import (
	"fmt"
	"github.com/aqatl/mal/dialog"
	"github.com/aqatl/mal/nyaa_scraper"
	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/urfave/cli"
	"math"
	"os/exec"
)

func malNyaaCui(ctx *cli.Context) error {
	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	entry := list.GetByID(cfg.SelectedID)
	if entry == nil {
		return fmt.Errorf("no entry found")
	}
	return startNyaaCui(cfg, entry.Title)
}

func alNyaaCui(ctx *cli.Context) error {
	al, err := loadAniList()
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	entry := al.GetMediaListById(cfg.ALSelectedID)
	if entry == nil {
		return fmt.Errorf("no entry found")
	}
	return startNyaaCui(cfg, entry.Title.Romaji)
}

func startNyaaCui(cfg *Config, searchTerm string) error {
	gui, err := gocui.NewGui(gocui.Output256)
	defer gui.Close()
	if err != nil {
		return fmt.Errorf("gocui error: %v", err)
	}
	nc := &nyaaCui{
		Cfg: cfg,

		SearchTerm: searchTerm,
		Category:   nyaa_scraper.AnimeEnglishTranslated,
		Filter:     nyaa_scraper.NoFilter,
	}
	gui.SetManager(nc)
	nc.setGuiKeyBindings(gui)

	gui.Cursor = false
	gui.Mouse = false
	gui.Highlight = true
	gui.SelFgColor = gocui.ColorGreen

	gui.Update(func(gui *gocui.Gui) error {
		nc.Reload(gui)
		return nil
	})

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

	SearchTerm string
	Category   nyaa_scraper.NyaaCategory
	Filter     nyaa_scraper.NyaaFilter

	Results     []nyaa_scraper.NyaaEntry
	MaxResults  int
	MaxPages    int
	LoadedPages int

	ResultsView *gocui.View
}

func (nc *nyaaCui) Layout(gui *gocui.Gui) error {
	w, h := gui.Size()
	if v, err := gui.SetView(ncInfoView, 0, 0, w-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Info"
		v.Editable = false

		fmt.Fprintf(v, "[%s]: displaying %d out of %d results",
			nc.SearchTerm, len(nc.Results), nc.MaxResults)
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

		//TODO Better/clearer results printing
		for _, result := range nc.Results {
			fmt.Fprintf(v, "%s %s %v %d %d %d\n",
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

		c := color.New(color.FgCyan).SprintFunc()
		fmt.Fprintln(v, c("d"), "download", c("l"), "load next page",
			c("c"), "category", c("f"), "filters")
	}

	return nil
}

func (nc *nyaaCui) Editor(gui *gocui.Gui) func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	//TODO it's too big
	return func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
		switch {
		case key == gocui.KeyArrowDown || ch == 'j':
			_, oy := v.Origin()
			_, y := v.Cursor()
			y += oy
			if y < len(nc.Results)-1 {
				v.MoveCursor(0, 1, false)
			}
		case key == gocui.KeyArrowUp || ch == 'k':
			v.MoveCursor(0, -1, false)
		case ch == 'd':
			_, y := v.Cursor()
			_, oy := v.Origin()
			y += oy
			if y >= len(nc.Results) {
				return
			}

			link := ""
			if entry := nc.Results[y]; entry.MagnetLink != "" {
				link = entry.MagnetLink
			} else if entry.TorrentLink != "" {
				link = entry.TorrentLink
			} else {
				dialog.JustShowOkDialog(gui, "Error", "No link found")
				return
			}

			if err := nc.Download(link); err != nil {
				gui.Update(func(gui *gocui.Gui) error {
					return err
				})
			}
		case ch == 'l':
			if nc.LoadedPages >= nc.MaxPages {
				return
			}
			nc.LoadedPages++
			go func() {
				resultPage, _ := nyaa_scraper.SearchSpecificPage(
					nc.SearchTerm,
					nyaa_scraper.AnimeEnglishTranslated,
					nyaa_scraper.NoFilter,
					nc.LoadedPages-1,
				)
				nc.Results = append(nc.Results, resultPage.Results...)
				gui.Update(func(gui *gocui.Gui) error {
					_, oy := nc.ResultsView.Origin()
					_, y := nc.ResultsView.Cursor()

					gui.DeleteView(ncInfoView)
					gui.DeleteView(ncResultsView)

					nc.Layout(gui)
					nc.ResultsView.SetOrigin(0, oy)
					nc.ResultsView.SetCursor(0, y)

					return nil
				})
			}()
		case ch == 'c':
			categories := make([]fmt.Stringer, len(nyaa_scraper.Categories))
			for i := range categories {
				categories[i] = nyaa_scraper.Categories[i]
			}
			selIdxChan, cleanUp, err := dialog.ListSelect(gui, "Select category", categories)
			if err != nil {
				gocuiReturnError(gui, err)
			}
			go func() {
				idx, ok := <-selIdxChan
				gui.Update(cleanUp)
				if ok {
					nc.Category = nyaa_scraper.Categories[idx]
					nc.Reload(gui)
				}
			}()
		case ch == 'f':
			filters := make([]fmt.Stringer, len(nyaa_scraper.Filters))
			for i := range filters {
				filters[i] = nyaa_scraper.Filters[i]
			}
			selIdxChan, cleanUp, err := dialog.ListSelect(gui, "Select filter", filters)
			if err != nil {
				gocuiReturnError(gui, err)
			}
			go func() {
				idx, ok := <-selIdxChan
				gui.Update(cleanUp)
				if ok {
					nc.Filter = nyaa_scraper.Filters[idx]
					nc.Reload(gui)
				}
			}()
		}
	}
}

func (nc *nyaaCui) Reload(gui *gocui.Gui) {
	var resultPage nyaa_scraper.NyaaResultPage
	var searchErr error
	f := func() {
		resultPage, searchErr = nyaa_scraper.Search(nc.SearchTerm, nc.Category, nc.Filter)
	}
	jobDone, err := dialog.StuffLoader(dialog.FitMessage(gui, "Loading "+nc.SearchTerm), f)
	if err != nil {
		gocuiReturnError(gui, err)
	}
	go func() {
		ok := <-jobDone
		if searchErr != nil {
			dialog.JustShowOkDialog(gui, "Error", searchErr.Error())
			return
		}
		if ok {
			nc.Results = resultPage.Results
			nc.MaxResults = resultPage.DisplayedOutOf
			nc.MaxPages = int(math.Ceil(float64(resultPage.DisplayedOutOf) /
				float64(resultPage.DisplayedTo-resultPage.DisplayedFrom+1)))
			nc.LoadedPages = 1
		}

		gui.Update(func(gui *gocui.Gui) error {
			gui.DeleteView(ncResultsView)
			gui.DeleteView(ncInfoView)
			return nil
		})
	}()
}

func (nc *nyaaCui) Download(link string) error {
	link = "\"" + link + "\""
	cmd := exec.Command(nc.Cfg.TorrentClientPath, nc.Cfg.TorrentClientArgs, link)
	cmd.Args = cmd.Args[1:] //Why they include app name in the arguments???
	return cmd.Start()
}

func (nc *nyaaCui) setGuiKeyBindings(gui *gocui.Gui) {
	gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitGocui)
}

func quitGocui(gui *gocui.Gui, view *gocui.View) error {
	return gocui.ErrQuit
}

func gocuiReturnError(gui *gocui.Gui, err error) {
	gui.Update(func(gui *gocui.Gui) error {
		return err
	})
}
