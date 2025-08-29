package pages

import (
	_ "embed"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/stonehenge-collective/html_builder/components"
	"github.com/stonehenge-collective/html_builder/draft"
	"github.com/stonehenge-collective/html_builder/page_builder"
	"github.com/stonehenge-collective/html_builder/shared_javascript"
)

//go:embed team.js
var teamJavascript string

func BuildTeamPage(d draft.Draft, u draft.User) string {
	var userNames, playerNames []string
	for _, u := range d.Users {
		userNames = append(userNames, html.EscapeString(u.DisplayName))
	}
	for _, p := range d.Players {
		playerNames = append(playerNames, html.EscapeString(p.FullName))
	}

	bootstrap := fmt.Sprintf("const userName = %s;\n", strconv.Quote(u.DisplayName))
	fullJS := shared_javascript.FirebaseInit + bootstrap + teamJavascript

	return page_builder.Build(page_builder.Page{
		Title: u.DisplayName+"'s Draft",
		Components: []page_builder.Component{
			components.Header{Title: u.DisplayName+"'s Draft"},
			components.Markup{HTML: `<p>Users: ` + strings.Join(userNames, ", ") + `</p>`},
			components.Markup{HTML: `<p>Players: ` + strings.Join(playerNames, ", ") + `</p>`},
			components.Script{JavaScript: fullJS},
		},
	})
}