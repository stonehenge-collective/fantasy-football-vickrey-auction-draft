package pages_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stonehenge-collective/html_builder/draft"
	"github.com/stonehenge-collective/html_builder/pages"
)

func TestBuildTeamPage(t *testing.T) {
	data, err := os.ReadFile("test_draft.json")
	if err != nil {
		t.Fatalf("read json: %v", err)
	}

	var draft draft.Draft
	if err := json.Unmarshal(data, &draft); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}

	html := pages.BuildTeamPage(draft, draft.Users[0])

	if err := os.WriteFile("team.html", []byte(html), 0o644); err != nil {
		t.Fatalf("write html: %v", err)
	}
}
