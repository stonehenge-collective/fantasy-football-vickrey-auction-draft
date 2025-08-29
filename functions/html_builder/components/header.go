package components

import "html"

type Header struct{ Title string }

func (h Header) Head() string   { return `` }
func (h Header) Body() string   { return `<h1>` + html.EscapeString(h.Title) + `</h1>` }
func (h Header) Script() string { return `` }