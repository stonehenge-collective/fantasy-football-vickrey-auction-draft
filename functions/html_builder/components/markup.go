package components

type Markup struct{ HTML string }

func (m Markup) Head() string   { return `` }
func (m Markup) Body() string   { return m.HTML  }
func (m Markup) Script() string { return `` }