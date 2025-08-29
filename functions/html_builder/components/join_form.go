package components

import (
	_ "embed"

	"github.com/stonehenge-collective/html_builder/shared_javascript"
)

//go:embed join_form.js
var javascript string

type JoinForm struct{}

func (JoinForm) Head() string {
	return `<style>
#joinDraftForm{display:flex;flex-direction:column;gap:.75rem;max-width:28rem}
#joinDraftForm input[type='text']{width:100%;padding:.5rem .75rem;font-size:1rem;box-sizing:border-box}
</style>`
}

func (JoinForm) Body() string {
	return `<form id="joinDraftForm">
<input type="text" id="sleeperUsername" name="sleeperUsername" placeholder="Enter Sleeper username" autocomplete="username" required>
<input type="text" id="draftPassword" name="draftPassword" placeholder="Password" autocomplete="current-password" required>
<button type="submit">Join Draft</button>
</form>`
}

func (JoinForm) Script() string {
	return `<script type="module">`+shared_javascript.FirebaseInit + javascript+`</script>`
}