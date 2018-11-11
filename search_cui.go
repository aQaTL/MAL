package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aqatl/mal/anilist"
	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/urfave/cli"
)

func alSearch(ctx *cli.Context) error {
	al, err := loadAniList(ctx)
	if err != nil {
		return nil
	}

	searchQuery := strings.TrimSpace(strings.Join(ctx.Args(), " "))
	results, err := anilist.Search(searchQuery, 1, 50, anilist.Anime, al.Token)
	if err != nil {
		return err
	}

	descriptionReplacer := strings.NewReplacer("<br>", "")
	for i := range results {
		results[i].Description = descriptionReplacer.Replace(results[i].Description)
	}

	gui, err := gocui.NewGui(gocui.Output256)
	defer gui.Close()
	if err != nil {
		return fmt.Errorf("gocui error: %v", err)
	}

	sc := &searchCui{
		Al:          al,
		Gui:         gui,
		SearchQuery: searchQuery,
		Results:     results,
		Mode:        scListView,
	}

	gui.SetManager(sc)
	sc.setGuiKeyBindings(gui)

	gui.Mouse = false
	gui.Highlight = true
	gui.Cursor = false
	gui.SelFgColor = gocui.ColorGreen

	if err = gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

const (
	scFiltersView   = "ncFiltersView"
	scShortcutsView = "scShortcutsView"
)

type searchCuiMode uint8

const (
	scListView searchCuiMode = iota
	scFullDetailsView
)

type searchCui struct {
	Al  *AniList
	Gui *gocui.Gui

	SearchQuery string
	Results     []anilist.MediaFull

	Mode   searchCuiMode
	SelIdx int
	Origin int
}

var searchResultHighlight = color.New(color.FgBlack, color.BgYellow)
var yellowC = color.New(color.FgYellow)
var cyanC = color.New(color.FgCyan)

func (sc *searchCui) Layout(gui *gocui.Gui) error {
	switch sc.Mode {
	case scListView:
		return sc.ListLayout()
	case scFullDetailsView:
		return sc.FullDetailsLayout()
	default:
		return fmt.Errorf("invalid mode: %d", sc.Mode)
	}
}

func (sc *searchCui) ListLayout() error {
	w, h := sc.Gui.Size()

	if err := sc.filtersView(); err != nil {
		return err
	}
	y := 4
	for i := sc.Origin; i < len(sc.Results) && y < h; i++ {
		result := &sc.Results[i]

		if v, err := sc.Gui.SetView(strconv.Itoa(result.Id), 0, y, w-1, y+7); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

			v.Frame = false
			v.Wrap = true
			v.Highlight = false
			v.Editable = true
			v.Editor = sc
			sc.Gui.SetViewOnTop(v.Name())

			if i == sc.SelIdx {
				searchResultHighlight.Fprintln(v, result.Title.UserPreferred)
			} else {
				yellowC.Fprintln(v, result.Title.UserPreferred)
			}
			cyanC.Fprint(v, strings.ToLower(fmt.Sprintf("%s | %s | %d eps | %s %d | %v\n",
				result.Format,
				result.Status,
				result.Episodes,
				result.Season,
				result.StartDate.Year,
				result.Genres,
			)))
			fmt.Fprintln(v, result.Description)

		}
		y += 6
	}

	if len(sc.Results) > 0 {
		sc.Gui.SetCurrentView(strconv.Itoa(sc.Results[0].Id))
	}

	return nil
}

func (sc *searchCui) FullDetailsLayout() error {
	w, h := sc.Gui.Size()

	if err := sc.filtersView(); err != nil {
		return err
	}

	if v, err := sc.Gui.SetView(strconv.Itoa(sc.Results[sc.SelIdx].Id), 0, 5, w-1, h-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Wrap = true

		fmt.Fprintln(v, sc.Results[sc.SelIdx].Description)
	}

	return nil
}

func (sc *searchCui) filtersView() error {
	w, _ := sc.Gui.Size()
	if v, err := sc.Gui.SetView(scFiltersView, 0, 0, w-1, 4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		fmt.Fprintln(v, "Search:", sc.SearchQuery)
		fmt.Fprintln(v, "Results:", len(sc.Results))
	}
	return nil
}

func (sc *searchCui) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch sc.Mode {
	case scListView:
		switch {
		case ch == 'j' || key == gocui.KeyArrowDown:
			sc.nextResult()
		case ch == 'k' || key == gocui.KeyArrowUp:
			sc.previousResult()
		case key == gocui.KeyEnter:
			if len(sc.Results) == 0 || sc.SelIdx < 0 || sc.SelIdx > len(sc.Results)-1 {
				return
			}
			sc.Mode = scFullDetailsView
			for _, result := range sc.Results {
				sc.Gui.DeleteView(strconv.Itoa(result.Id))
			}
		}
	case scFullDetailsView:
		switch {
		case ch == 'j' || key == gocui.KeyArrowDown:
			sc.nextResult()
		case ch == 'k' || key == gocui.KeyArrowUp:
			sc.previousResult()
		case key == gocui.KeyEnter:
			sc.Mode = scListView
			sc.Gui.DeleteView(strconv.Itoa(sc.Results[sc.SelIdx].Id))
		}
	}
}

func (sc *searchCui) nextResult() {
	if sc.SelIdx != len(sc.Results)-1 {
		sc.SelIdx++
		if _, h := sc.Gui.Size(); sc.SelIdx > (sc.Origin + int((h-6)/6)) {
			sc.Origin++
		}
		sc.Gui.DeleteView(strconv.Itoa(sc.Results[sc.SelIdx].Id))
		sc.Gui.DeleteView(strconv.Itoa(sc.Results[sc.SelIdx-1].Id))
	}
}

func (sc *searchCui) previousResult() {
	if sc.SelIdx != 0 {
		sc.SelIdx--
		if sc.SelIdx < sc.Origin {
			sc.Origin--
		}
		sc.Gui.DeleteView(strconv.Itoa(sc.Results[sc.SelIdx].Id))
		sc.Gui.DeleteView(strconv.Itoa(sc.Results[sc.SelIdx+1].Id))
	}
}

func (sc *searchCui) setGuiKeyBindings(gui *gocui.Gui) {
	gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitGocui)
}
