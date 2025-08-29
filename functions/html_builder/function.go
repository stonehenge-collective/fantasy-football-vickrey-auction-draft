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

	if _, err := client.Collection("drafts").
		Doc(draftID).
		Collection("pages").
		Doc("public").
		Set(ctx, map[string]any{"html": publicHtml}); err != nil {

		return fmt.Errorf("write public page: %w", err)
	}

	for _, user := range draft.Users {
		teamHtml := pages.BuildTeamPage(draft, user)
		if _, err := client.Collection("drafts").
			Doc(draftID).
			Collection("pages").
			Doc(user.DisplayName).
			Set(ctx, map[string]any{"html": teamHtml}); err != nil {

			return fmt.Errorf("write team page for %s: %w", user.DisplayName, err)
		}
	}

	return nil
}


