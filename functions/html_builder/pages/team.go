package pages

import (
	"html"
	"strings"

	"github.com/stonehenge-collective/html_builder/components"
	"github.com/stonehenge-collective/html_builder/draft"
	"github.com/stonehenge-collective/html_builder/page_builder"
)

func BuildTeamPage(d draft.Draft) string {
	var userNames, playerNames []string
	for _, u := range d.Users {
		userNames = append(userNames, html.EscapeString(u.DisplayName))
	}
	for _, p := range d.Players {
		playerNames = append(playerNames, html.EscapeString(p.FullName))
	}

	return page_builder.Build(page_builder.Page{
		Title: "Johnor's Draft",
		Components: []page_builder.Component{
			components.Header{Title: "Join Draft"},
			components.JoinForm{},
			components.Markup{HTML: `<p>Users: ` + strings.Join(userNames, ", ") + `</p>`},
			components.Markup{HTML: `<p>Players: ` + strings.Join(playerNames, ", ") + `</p>`},
		},
	})
}