package main

import (
	"fmt"
	"math"
	"os/exec"
	"strings"

	"github.com/aqatl/mal/dialog"
	"github.com/aqatl/mal/nyaa_scraper"
	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/urfave/cli"
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
	return startNyaaCui(
		cfg,
		entry.Title,
		fmt.Sprintf("%s %d/%d", entry.Title, entry.WatchedEpisodes, entry.Episodes),
	)
}

func alNyaaCui(ctx *cli.Context) error {
	al, err := loadAniList(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	entry := al.GetMediaListById(cfg.ALSelectedID)
	if entry == nil {
		return fmt.Errorf("no entry found")
	}

	searchTerm := entry.Title.UserPreferred
	if ctx.Bool("alt") {
		fmt.Printf("Select desired title\n\n")
		if searchTerm = chooseStrFromSlice(sliceOfEntryTitles(entry)); searchTerm == "" {
			return fmt.Errorf("no alternative titles")
		}
	} else if ctx.NArg() > 0 {
		searchTerm = strings.Join(ctx.Args(), " ")
	}

	return startNyaaCui(
		cfg,
		searchTerm,
		fmt.Sprintf("%s %d/%d", searchTerm, entry.Progress, entry.Episodes),
	)
}

func startNyaaCui(cfg *Config, searchTerm, displayedInfo string) error {
	gui, err := gocui.NewGui(gocui.Output256)
	defer gui.Close()
	if err != nil {
		return fmt.Errorf("gocui error: %v", err)
	}
	nc := &nyaaCui{
		Gui: gui,
		Cfg: cfg,

		SearchTerm:    searchTerm,
		DisplayedInfo: displayedInfo,
		Category:      nyaa_scraper.AnimeEnglishTranslated,
		Filter:        nyaa_scraper.TrustedOnly,
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
	Gui *gocui.Gui
	Cfg *Config

	SearchTerm    string
	DisplayedInfo string
	Category      nyaa_scraper.NyaaCategory
	Filter        nyaa_scraper.NyaaFilter

	Results     []nyaa_scraper.NyaaEntry
	MaxResults  int
	MaxPages    int
	LoadedPages int

	ResultsView *gocui.View
}

var red = color.New(color.FgRed).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()

func (nc *nyaaCui) Layout(gui *gocui.Gui) error {
	w, h := gui.Size()
	if v, err := gui.SetView(ncInfoView, 0, 0, w-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Info"
		v.Editable = false

		fmt.Fprintf(v, "[%s]: displaying %d out of %d results",
			nc.DisplayedInfo, len(nc.Results), nc.MaxResults)
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
		v.Editor = gocui.EditorFunc(nc.GetEditor())

		gui.SetCurrentView(ncResultsView)
		nc.ResultsView = v

		//TODO Better/clearer results printing
		for _, result := range nc.Results {
			fmt.Fprintln(v,
				result.Title,
				red(result.Size),
				cyan(result.DateAdded.Format("15:04:05 02-01-2006 MST")),
				green(result.Seeders),
				red(result.Leechers),
				blue(result.CompletedDownloads),
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

func (nc *nyaaCui) GetEditor() func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
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
		case ch == 'g':
			v.SetCursor(0, 0)
			v.SetOrigin(0, 0)
		case ch == 'G':
			_, viewH := v.Size()
			totalH := len(nc.Results)
			if totalH <= viewH {
				v.SetCursor(0, totalH-1)
			} else {
				v.SetOrigin(0, totalH-viewH)
				v.SetCursor(0, viewH-1)
			}
		case ch == 'd':
			_, y := v.Cursor()
			_, oy := v.Origin()
			y += oy
			nc.Download(y)
		case ch == 'l':
			nc.LoadNextPage()
		case ch == 'c':
			nc.ChangeCategory()
		case ch == 'f':
			nc.ChangeFilter()
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

func (nc *nyaaCui) Download(yIdx int) {
	if yIdx >= len(nc.Results) {
		return
	}

	link := ""
	if entry := nc.Results[yIdx]; entry.MagnetLink != "" {
		link = entry.MagnetLink
	} else if entry.TorrentLink != "" {
		link = entry.TorrentLink
	} else {
		dialog.JustShowOkDialog(nc.Gui, "Error", "No link found")
		return
	}

	link = "\"" + link + "\""
	cmd := exec.Command(nc.Cfg.TorrentClientPath, nc.Cfg.TorrentClientArgs...)
	cmd.Args = append(cmd.Args, link)
	cmd.Args = cmd.Args[1:] //Why they include app name in the arguments???
	if err := cmd.Start(); err != nil {
		gocuiReturnError(nc.Gui, err)
	}
}

func (nc *nyaaCui) LoadNextPage() {
	if nc.LoadedPages >= nc.MaxPages {
		return
	}
	nc.LoadedPages++
	go func() {
		resultPage, _ := nyaa_scraper.SearchSpecificPage(
			nc.SearchTerm,
			nc.Category,
			nc.Filter,
			nc.LoadedPages,
		)
		nc.Results = append(nc.Results, resultPage.Results...)
		nc.Gui.Update(func(gui *gocui.Gui) error {
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
}

func (nc *nyaaCui) ChangeCategory() {
	selIdxChan, cleanUp, err := dialog.ListSelect(nc.Gui, "Select category", nyaa_scraper.Categories)
	if err != nil {
		gocuiReturnError(nc.Gui, err)
	}
	go func() {
		idx, ok := <-selIdxChan
		nc.Gui.Update(cleanUp)
		if ok {
			nc.Category = nyaa_scraper.Categories[idx]
			nc.Reload(nc.Gui)
		}
	}()
}

func (nc *nyaaCui) ChangeFilter() {
	selIdxChan, cleanUp, err := dialog.ListSelect(nc.Gui, "Select filter", nyaa_scraper.Filters)
	if err != nil {
		gocuiReturnError(nc.Gui, err)
	}
	go func() {
		idx, ok := <-selIdxChan
		nc.Gui.Update(cleanUp)
		if ok {
			nc.Filter = nyaa_scraper.Filters[idx]
			nc.Reload(nc.Gui)
		}
	}()
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
