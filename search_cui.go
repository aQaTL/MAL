package main

import (
	"fmt"
	"github.com/aqatl/mal/anilist"
	"github.com/jroimartin/gocui"
	"github.com/urfave/cli"
	"strconv"
	"strings"
	"github.com/fatih/color"
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

type searchCui struct {
	Al  *AniList
	Gui *gocui.Gui

	SearchQuery string
	Results     []anilist.MediaFull

	SelIdx int
	Origin int
}

var searchResultHighlight = color.New(color.FgBlack, color.BgYellow)
var yellowC = color.New(color.FgYellow)
var cyanC = color.New(color.FgCyan)

func (sc *searchCui) Layout(gui *gocui.Gui) error {
	w, h := gui.Size()

	if v, err := gui.SetView(scFiltersView, 0, 0, w-1, 4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		fmt.Fprintln(v, "Search:", sc.SearchQuery)
		fmt.Fprintln(v, "Results:", len(sc.Results))
	}

	descriptionReplacer := strings.NewReplacer("<br>", "", "\n", " ")
	y := 4
	for i := sc.Origin; i < len(sc.Results) && y < h; i++ {
		result := &sc.Results[i]

		if v, err := gui.SetView(strconv.Itoa(result.Id), 0, y, w-1, y+7); err != nil {
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
			fmt.Fprintln(v, descriptionReplacer.Replace(result.Description))

		}
		y += 6
	}

	if len(sc.Results) > 0 {
		gui.SetCurrentView(strconv.Itoa(sc.Results[0].Id))
	}

	return nil
}

func (sc *searchCui) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch == 'j' || key == gocui.KeyArrowDown:
		if sc.SelIdx != len(sc.Results)-1 {
			sc.SelIdx++
			if _, h := sc.Gui.Size(); sc.SelIdx > (sc.Origin + int((h-6)/6)) {
				sc.Origin++
			}
			sc.Gui.DeleteView(strconv.Itoa(sc.Results[sc.SelIdx].Id))
			sc.Gui.DeleteView(strconv.Itoa(sc.Results[sc.SelIdx-1].Id))
		}
	case ch == 'k' || key == gocui.KeyArrowUp:
		if sc.SelIdx != 0 {
			sc.SelIdx--
			if sc.SelIdx < sc.Origin {
				sc.Origin--
			}
			sc.Gui.DeleteView(strconv.Itoa(sc.Results[sc.SelIdx].Id))
			sc.Gui.DeleteView(strconv.Itoa(sc.Results[sc.SelIdx+1].Id))
		}
	}
}

func (sc *searchCui) setGuiKeyBindings(gui *gocui.Gui) {
	gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitGocui)
}
