package page_builder

import (
	"bytes"
	"fmt"
)

type Component interface {
	Head() string   // <style> or <script> tags that belong in <head>
	Body() string   // HTML that belongs in <body>
	Script() string // late-loaded <script> tags (put just before </body>)
}

type Page struct {
	Title      string
	Components [] Component
}

func Build(p Page) string {
	var head, body, script bytes.Buffer

	for _, c := range p.Components {
		head.WriteString(c.Head())
		body.WriteString(c.Body())
		script.WriteString(c.Script())
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>%s</title>
%s
</head>
<body>
<div id="root">
%s
</div>
%s
</body>
</html>`, p.Title, head.String(), body.String(), script.String())
}
