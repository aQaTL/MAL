package main

import (
	"fmt"
	"math"
	"os/exec"
	"strings"

	"regexp"

	"sort"

	"github.com/aqatl/mal/dialog"
	ns "github.com/aqatl/mal/nyaa_scraper"
	"github.com/atotto/clipboard"
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
		cfg.NyaaQuality,
	)
}

type NyaaAlt struct {
	Query string
	Id    int
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

	if alt := ctx.String("custom"); alt != "" {
		addCustomAlt(alt+" "+strings.Join(ctx.Args(), " "), cfg)
		return nil
	}

	var customAlt *string = nil
	for _, alt := range cfg.NyaaAlts {
		if alt.Id == entry.Id {
			customAlt = &alt.Query
			break
		}
	}

	searchTerm := entry.Title.UserPreferred
	if ctx.Bool("alt") {
		fmt.Printf("Select desired title\n\n")
		alts := sliceOfEntryTitles(entry)
		if customAlt != nil {
			alts = append(alts, *customAlt)
		}
		if searchTerm = chooseStrFromSlice(alts); searchTerm == "" {
			return fmt.Errorf("no alternative titles")
		}
	} else if ctx.NArg() > 0 {
		searchTerm = strings.Join(ctx.Args(), " ")
	} else if customAlt != nil {
		searchTerm = *customAlt
	}

	if err := startNyaaCui(
		cfg,
		searchTerm,
		fmt.Sprintf("%s %d/%d", searchTerm, entry.Progress, entry.Episodes),
		cfg.NyaaQuality,
	); err != nil {
		return err
	}

	alPrintEntryDetails(entry)
	return nil
}

func addCustomAlt(newAlt string, cfg *Config) {
	// Assumes the entry ID is valid
	defer cfg.Save()
	for i, alt := range cfg.NyaaAlts {
		if alt.Id == cfg.ALSelectedID {
			cfg.NyaaAlts[i].Query = newAlt
			return
		}
	}

	cfg.NyaaAlts = append(cfg.NyaaAlts, NyaaAlt{
		Id:    cfg.ALSelectedID,
		Query: newAlt,
	})
}

func startNyaaCui(cfg *Config, searchTerm, displayedInfo, quality string) error {
	gui, err := gocui.NewGui(gocui.Output256)
	defer gui.Close()
	if err != nil {
		return fmt.Errorf("gocui error: %v", err)
	}

	qualityRe, err := regexp.Compile(regexp.QuoteMeta(quality))
	if err != nil {
		return fmt.Errorf("failed to parse your quality tag")
	}

	nc := &nyaaCui{
		Gui: gui,
		Cfg: cfg,

		SearchTerm:    searchTerm,
		DisplayedInfo: displayedInfo,
		Category:      ns.AnimeEnglishTranslated,
		Filter:        ns.TrustedOnly,

		QualityFilter: qualityRe,
	}
	gui.SetManager(nc)
	nc.setGuiKeyBindings(gui)

	gui.Cursor = false
	gui.Mouse = false
	gui.Highlight = true
	gui.SelFgColor = gocui.ColorGreen

	gui.Update(func(gui *gocui.Gui) error {
		nc.Reload()
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
	Category      ns.NyaaCategory
	Filter        ns.NyaaFilter

	Results     []ns.NyaaEntry
	MaxResults  int
	MaxPages    int
	LoadedPages int

	TitleFilter   *regexp.Regexp
	QualityFilter *regexp.Regexp

	ResultsView      *gocui.View
	DisplayedIndexes []int
}

var red = color.New(color.FgRed).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()

var boldRed = color.New(color.FgRed).Add(color.Bold).SprintFunc()
var boldGreen = color.New(color.FgGreen).Add(color.Bold).SprintFunc()
var boldYellow = color.New(color.FgYellow).Add(color.Bold).SprintFunc()

func (nc *nyaaCui) Layout(gui *gocui.Gui) error {
	w, h := gui.Size()

	if v, err := gui.SetView(ncResultsView, 0, 3, w-1, h-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Search results"
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = true
		v.Editable = true
		v.Editor = gocui.EditorFunc(nc.GetEditor())

		gui.SetCurrentView(ncResultsView)
		nc.ResultsView = v

		// TODO Better/clearer results printing
		nc.DisplayedIndexes = make([]int, 0, len(nc.Results))
		for i, result := range nc.Results {
			if nc.TitleFilter != nil && !nc.TitleFilter.MatchString(result.Title) {
				continue
			}
			if nc.QualityFilter != nil && !nc.QualityFilter.MatchString(result.Title) {
				continue
			}

			title := result.Title
			switch result.Class {
			case ns.Default:
				title = boldYellow(title)
			case ns.Trusted:
				title = boldGreen(title)
			case ns.Danger:
				title = boldRed(title)
			}

			fmt.Fprintln(v,
				title,
				red(result.Size),
				cyan(result.DateAdded.Format("15:04 02-01-2006")),
				green(result.Seeders),
				red(result.Leechers),
				blue(result.CompletedDownloads),
			)
			nc.DisplayedIndexes = append(nc.DisplayedIndexes, i)
		}
	}

	if v, err := gui.SetView(ncInfoView, 0, 0, w-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Info"
		v.Editable = false

		fmt.Fprintf(v, "[%s]: displaying %d out of %d results",
			nc.DisplayedInfo, len(nc.DisplayedIndexes), nc.MaxResults)
	}

	if v, err := gui.SetView(ncShortcutsView, 0, h-3, w-1, h-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Shortcuts"
		v.Editable = false

		c := color.New(color.FgCyan).SprintFunc()
		fmt.Fprintln(v,
			c("d"), "download",
			c("D"), "copy torrent link",
			c("l"), "load next page",
			c("c"), "category",
			c("f"), "filters",
			c("t"), "tags",
			c("p"), "quality",
			c("r"), "reload",
		)
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
			if y < len(nc.DisplayedIndexes)-1 {
				v.MoveCursor(0, 1, false)
			}
		case key == gocui.KeyArrowUp || ch == 'k':
			v.MoveCursor(0, -1, false)
		case ch == 'g':
			v.SetCursor(0, 0)
			v.SetOrigin(0, 0)
		case ch == 'G':
			_, viewH := v.Size()
			totalH := len(nc.DisplayedIndexes)
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
		case ch == 't':
			nc.FilterByTag()
		case ch == 'p':
			nc.FilterByQuality()
		case ch == 'r':
			nc.Reload()
		case ch == 'D':
			_, y := v.Cursor()
			_, oy := v.Origin()
			y += oy
			nc.CopyLinkToClipboard(y)
		}
	}
}

func (nc *nyaaCui) Reload() {
	var resultPage ns.NyaaResultPage
	var searchErr error
	f := func() {
		resultPage, searchErr = ns.Search(nc.SearchTerm, nc.Category, nc.Filter)
	}
	jobDone, err := dialog.StuffLoader(dialog.FitMessage(nc.Gui, "Loading "+nc.SearchTerm), f)
	if err != nil {
		gocuiReturnError(nc.Gui, err)
	}
	go func() {
		ok := <-jobDone
		if searchErr != nil {
			dialog.JustShowOkDialog(nc.Gui, "Error", searchErr.Error())
			return
		}
		if ok {
			nc.Results = resultPage.Results
			nc.MaxResults = resultPage.DisplayedOutOf
			nc.MaxPages = int(math.Ceil(float64(resultPage.DisplayedOutOf) /
				float64(resultPage.DisplayedTo-resultPage.DisplayedFrom+1)))
			nc.LoadedPages = 1
		}

		nc.Gui.Update(func(gui *gocui.Gui) error {
			gui.DeleteView(ncResultsView)
			gui.DeleteView(ncInfoView)
			return nil
		})
	}()
}

func (nc *nyaaCui) Download(yIdx int) {
	if yIdx >= len(nc.DisplayedIndexes) {
		return
	}

	link := ""
	if entry := nc.Results[nc.DisplayedIndexes[yIdx]]; entry.MagnetLink != "" {
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
	if len(nc.Cfg.TorrentClientArgs) > 0 {
		cmd.Args = cmd.Args[1:] // Why they include app name in the arguments???
	}
	if err := cmd.Start(); err != nil {
		gocuiReturnError(nc.Gui, err)
	}
}

func (nc *nyaaCui) CopyLinkToClipboard(yIdx int) {
	if yIdx >= len(nc.DisplayedIndexes) {
		return
	}

	link := ""
	if entry := nc.Results[nc.DisplayedIndexes[yIdx]]; entry.MagnetLink != "" {
		link = entry.MagnetLink
	} else if entry.TorrentLink != "" {
		link = entry.TorrentLink
	} else {
		dialog.JustShowOkDialog(nc.Gui, "Error", "No link found")
		return
	}

	if err := clipboard.WriteAll(link); err == nil {
		dialog.JustShowOkDialog(nc.Gui, "Clipboard", "Link copied into clipboard")
	}
}

func (nc *nyaaCui) LoadNextPage() {
	if nc.LoadedPages >= nc.MaxPages {
		return
	}
	nc.LoadedPages++
	go func() {
		resultPage, _ := ns.SearchSpecificPage(
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
	selIdxChan, cleanUp, err := dialog.ListSelect(nc.Gui, "Select category", ns.Categories, false)
	if err != nil {
		gocuiReturnError(nc.Gui, err)
	}
	go func() {
		idxs, ok := <-selIdxChan
		nc.Gui.Update(cleanUp)
		if ok {
			nc.Category = ns.Categories[idxs[0]]
			nc.Reload()
		}
	}()
}

func (nc *nyaaCui) ChangeFilter() {
	selIdxChan, cleanUp, err := dialog.ListSelect(nc.Gui, "Select filter", ns.Filters, false)
	if err != nil {
		gocuiReturnError(nc.Gui, err)
	}
	go func() {
		idxs, ok := <-selIdxChan
		nc.Gui.Update(cleanUp)
		if ok {
			nc.Filter = ns.Filters[idxs[0]]
			nc.Reload()
		}
	}()
}

var tagRegex = `(?U)\[(.+)\]`

func (nc *nyaaCui) FilterByTag() {
	tags := make([]string, 1, len(nc.Results)+1)
	tagsDup := make(map[string]struct{})
	re := regexp.MustCompile(tagRegex)
	for _, result := range nc.Results {
		if tsm := re.FindStringSubmatch(result.Title); len(tsm) >= 2 && tsm[1] != "" {
			if _, ok := tagsDup[tsm[1]]; !ok {
				tags = append(tags, tsm[1])
				tagsDup[tsm[1]] = struct{}{}
			}
		}
	}
	sort.Strings(tags)
	tags[0] = "None"

	selIdxChan, cleanUp, err := dialog.ListSelect(nc.Gui, "Select tag filter", tags, true)
	if err != nil {
		gocuiReturnError(nc.Gui, err)
	}
	go func() {
		idxs, ok := <-selIdxChan
		nc.Gui.Update(cleanUp)
		if ok {
			containsNone := false
			for _, v := range idxs {
				if v == 0 {
					containsNone = true
					break
				}
			}
			if len(idxs) == 0 || containsNone {
				nc.TitleFilter = nil
			} else {
				tagBuffer := strings.Builder{}
				for i, v := range idxs {
					tagBuffer.WriteString("\\[")
					tagBuffer.WriteString(regexp.QuoteMeta(tags[v]))
					tagBuffer.WriteString("\\]")
					if i < len(idxs)-1 {
						tagBuffer.WriteString("|")
					}
				}

				regex, err := regexp.Compile(tagBuffer.String())
				if err != nil {
					gocuiReturnError(nc.Gui, err)
				}
				nc.TitleFilter = regex
			}
			nc.Gui.Update(func(gui *gocui.Gui) error {
				gui.DeleteView(ncInfoView)
				gui.DeleteView(ncResultsView)
				return nil
			})
		}
	}()
}

var qualityTagRegex = `(\d{3,4}p)`

func (nc *nyaaCui) FilterByQuality() {
	tags := make([]string, 1, len(nc.Results)+1)
	tagsDup := make(map[string]struct{})
	re := regexp.MustCompile(qualityTagRegex)
	for _, result := range nc.Results {
		if tsm := re.FindStringSubmatch(result.Title); len(tsm) >= 2 && tsm[1] != "" {
			if _, ok := tagsDup[tsm[1]]; !ok {
				tags = append(tags, tsm[1])
				tagsDup[tsm[1]] = struct{}{}
			}
		}
	}
	sort.Strings(tags)
	tags[0] = "None"

	selIdxChan, cleanUp, err := dialog.ListSelect(nc.Gui, "Select quality filter", tags, true)
	if err != nil {
		gocuiReturnError(nc.Gui, err)
	}
	go func() {
		idxs, ok := <-selIdxChan
		nc.Gui.Update(cleanUp)
		if ok {
			containsNone := false
			for _, v := range idxs {
				if v == 0 {
					containsNone = true
					break
				}
			}
			if len(idxs) == 0 || containsNone {
				nc.QualityFilter = nil
			} else {
				tagBuffer := strings.Builder{}
				for i, v := range idxs {
					tagBuffer.WriteString(regexp.QuoteMeta(tags[v]))
					if i < len(idxs)-1 {
						tagBuffer.WriteString("|")
					}
				}

				regex, err := regexp.Compile(tagBuffer.String())
				if err != nil {
					gocuiReturnError(nc.Gui, err)
				}
				nc.QualityFilter = regex
			}
			nc.Gui.Update(func(gui *gocui.Gui) error {
				gui.DeleteView(ncInfoView)
				gui.DeleteView(ncResultsView)
				return nil
			})
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
