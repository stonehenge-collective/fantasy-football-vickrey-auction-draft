package function

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
	"github.com/stonehenge-collective/html_builder/draft"
	"github.com/stonehenge-collective/html_builder/pages"
	"google.golang.org/protobuf/proto"
)

// constant HTML template; the three %s placeholders are:
//
//   1. <div id="root">…</div>            – generated users & players
//   2. draftId used by mainDocRef
//   3. draftId used by teamPageRef
//
const pageTemplate = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Join Draft</title>
  </head>
  <style>
    #joinDraftForm {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
      max-width: 28rem;
    }
    #joinDraftForm input[type='text'] {
      width: 100%%;
      padding: 0.5rem 0.75rem;
      font-size: 1rem;
      box-sizing: border-box;
    }
  </style>
  <body>
    <div id="root">%s</div>
    <form id="joinDraftForm">
      <input
        type="text"
        id="sleeperUsername"
        name="sleeperUsername"
        placeholder="Enter Sleeper username to join draft."
        title="Enter Sleeper username to join draft."
        autocomplete="username"
        required
      />
      <input
        type="text"
        id="draftPassword"
        name="draftPassword"
        placeholder="Make up a draft password or, if you're re-joining, use your existing password."
        title="Make up a draft password or, if you're re-joining, use your existing password."
        autocomplete="current-password"
        required
      />
      <button type="submit">Join Draft</button>
    </form>

    <script type="module">
      
    </script>
  </body>
</html>
`

// HandleDraft regenerates the public page every time a draft document is created or updated.
func HandleDraft(ctx context.Context, e event.Event) error {
	// Ignore deletes.
	if e.Type() == "google.cloud.firestore.document.v1.deleted" {
		return nil
	}

	// ─────────────── Firestore event payload ───────────────
	var data firestoredata.DocumentEventData

	opts := proto.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal(e.Data(), &data); err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}
	if data.GetValue() == nil {
		return errors.New(`invalid message: "Value" not present`)
	}

	// ─────────────── Path parsing ───────────────
	const prefix = "/documents/"
	name := data.GetValue().GetName()
	idx := strings.Index(name, prefix)
	if idx == -1 {
		return fmt.Errorf("invalid resource name: %s", name)
	}
	relPath := name[idx+len(prefix):]                    // e.g. drafts/0776cbd3…
	draftID := relPath[strings.LastIndex(relPath, "/")+1:]

	// ─────────────── Firestore read ───────────────
	client, err := firestore.NewClient(ctx, "test-vickrey")
	if err != nil {
		return fmt.Errorf("firestore client: %w", err)
	}
	defer client.Close()

	snap, err := client.Doc(relPath).Get(ctx)
	if err != nil {
		return fmt.Errorf("read draft %s: %w", relPath, err)
	}

	var draft draft.Draft
	if err := snap.DataTo(&draft); err != nil {
		return fmt.Errorf("decode draft: %w", err)
	}

	publicHtml := pages.BuildPublicPage(draft)

	// ─────────────── Firestore write ───────────────
	if _, err := client.Collection("drafts").
		Doc(draftID).
		Collection("pages").
		Doc("public").
		Set(ctx, map[string]any{"html": publicHtml}); err != nil {

		return fmt.Errorf("write public page: %w", err)
	}

	return nil
}


