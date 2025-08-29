package components

type Script struct{ JavaScript string }

func (s Script) Head() string   { return `` }
func (s Script) Body() string   { return ``  }
func (s Script) Script() string { return `<script type="module">`+s.JavaScript+`</script>` }