package main

import (
	"github.com/aqatl/mal/mal"
	"math"
	"text/template"
)

const PrettyListTemplate = `No{{printf "%57s" "Title"}}{{printf "%8s" "Eps"}}{{printf "%6s" "Score"}}{{printf "%7s" "ID"}}
================================================================================{{range $index, $var := .List}}
{{if eq .ID $.SelectedID}}{{"\033[93;1m"}}{{end}}{{len $.List | minus $index | abs | printf "%2d"}}{{.Title | printf " %56.56s"}}{{printf "%d/%d" .WatchedEpisodes .Episodes | printf "%8s"}}{{.MyScore | printf "%6d"}}{{.ID | printf "%7d"}}{{if eq .ID $.SelectedID}}{{"\033[0m "}}{{end}}{{end}}
`

var PrettyList = template.Must(
	template.New("prettyList").
		Funcs(template.FuncMap{
			"plus": func(a, b int) int {
				return a + b
			},
			"minus": func(a, b int) int {
				return a - b
			},
			"abs": func(a int) int {
				return int(math.Abs(float64(a)))
			},
		}).
		Parse(PrettyListTemplate),
)

type PrettyListData struct {
	List       []*mal.Anime
	SelectedID int
}
