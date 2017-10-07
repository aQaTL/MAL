package main

import (
	"math"
	"text/template"
)

const PrettyListTemplate = `No{{printf "%40s" "Title"}}{{printf "%8s" "Eps"}}{{printf "%6s" "Score"}}
========================================================{{range $index, $var := .}}
{{len $ | minus $index | abs | printf "%2d"}}{{.Title | printf "%40s"}}{{printf "%d/%d" .WatchedEpisodes .Episodes | printf "%8s"}}{{.MyScore | printf "%6d"}}{{end}}
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
