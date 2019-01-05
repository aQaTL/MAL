package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aqatl/mal/anilist"
	"github.com/aqatl/mal/dialog"
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

	gui, err := gocui.NewGui(gocui.OutputNormal)
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
	scSearchView    = "scSearchView"
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
var yellowC = color.New(color.FgYellow, color.Bold)
var cyanC = color.New(color.FgCyan, color.Bold)

// Safe to call from another goroutine
func (sc *searchCui) reload() {
	results, err := anilist.Search(sc.SearchQuery, 1, 50, anilist.Anime, sc.Al.Token)
	if err != nil {
		dialog.JustShowOkDialog(sc.Gui, "Error",
			strings.TrimSpace(strings.Replace(err.Error(), "\n", " ", -1)))
	}
	sc.Gui.Update(func(gui *gocui.Gui) error {
		sc.Results = results
		sc.Gui.SetManager(sc)
		sc.setGuiKeyBindings(sc.Gui)
		return nil
	})
}

func (sc *searchCui) setGuiKeyBindings(gui *gocui.Gui) {
	gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitGocui)
}

func (sc *searchCui) Layout(gui *gocui.Gui) error {
	switch sc.Mode {
	case scListView:
		return sc.listLayout()
	case scFullDetailsView:
		return sc.fullDetailsLayout()
	default:
		return fmt.Errorf("invalid mode: %d", sc.Mode)
	}
}

func (sc *searchCui) listLayout() error {
	w, h := sc.Gui.Size()
	h -= 4

	if err := sc.filtersView(); err != nil {
		return err
	}
	if err := sc.searchView(); err != nil {
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
				searchResultHighlight.Fprintf(v, "%s (%s)\n", result.Title.Romaji, result.Title.English)
			} else {
				yellowC.Fprintf(v, "%s (%s)\n", result.Title.Romaji, result.Title.English)
			}
			cyanC.Fprint(v, strings.ToLower(
				fmt.Sprintf("%s | %s | %d eps | %s %d | %d%% | %v\n",
					result.Format,
					result.Status,
					result.Episodes,
					result.Season,
					result.StartDate.Year,
					result.AverageScore,
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

func (sc *searchCui) fullDetailsLayout() error {
	w, h := sc.Gui.Size()

	if err := sc.filtersView(); err != nil {
		return err
	}

	if v, err := sc.Gui.SetView(strconv.Itoa(sc.Results[sc.SelIdx].Id), 0, 5, w-1, h-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Wrap = true
		v.Editor = sc
		v.Editable = true

		sc.Gui.SetCurrentView(v.Name())

		fmt.Fprintln(v, sc.Results[sc.SelIdx].Description)
	}

	return nil
}

func (sc *searchCui) filtersView() error {
	w, _ := sc.Gui.Size()
	v, err := sc.Gui.SetView(scFiltersView, 0, 0, w-1, 4)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Editor = sc

		fmt.Fprintln(v, "Search:", sc.SearchQuery)
		fmt.Fprintln(v, "Results:", len(sc.Results))
	}

	return nil
}

func (sc *searchCui) searchView() error {
	w, h := sc.Gui.Size()
	v, err := sc.Gui.SetView(scSearchView, 0, h-3, w-1, h-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Search"
		v.Frame = true
		v.Editor = sc.searchViewEditor()

		fmt.Fprint(v, sc.SearchQuery)
	}

	return nil
}

func (sc *searchCui) searchViewEditor() gocui.Editor {
	return gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
		gocui.DefaultEditor.Edit(v, key, ch, mod)
		go sc.reload()
	})
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
		case ch == 'a':
			if len(sc.Results) == 0 {
				return
			}
			sc.addEntry()
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

func (sc *searchCui) addEntry() {
	if entry := sc.Al.GetMediaListById(sc.Results[sc.SelIdx].Id); entry != nil {
		dialog.JustShowOkDialog(sc.Gui, "Add entry",
			"Entry already added (on list "+entry.Status.String()+")")
		return
	}

	entry, err := anilist.AddMediaListEntry(sc.Results[sc.SelIdx].Id, anilist.Planning, sc.Al.Token)
	if err != nil {
		dialog.JustShowOkDialog(sc.Gui, "Error", err.Error())
		return
	}
	dialog.JustShowOkDialog(sc.Gui, "Success", entry.Title.UserPreferred+" added")

	sc.Al.List = append(sc.Al.List, entry)
	sc.Gui.Update(func(gui *gocui.Gui) error {
		return saveAniListAnimeLists(sc.Al)
	})
}
