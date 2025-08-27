package function

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestHandleDraft(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:9090")
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-vickrey")

	ctx := context.Background()
	if _, err := firestore.NewClient(ctx, "test-vickrey"); err != nil {
		t.Fatalf("firestore client: %v", err)
	}

	e := cloudevents.NewEvent()
	e.SetID("test-event-id")
	e.SetSource("//firestore.googleapis.com/projects/test-vickrey/databases/(default)")
	e.SetType("google.cloud.firestore.document.v1.written")
	e.SetTime(time.Now())
	e.SetDataContentType(cloudevents.ApplicationJSON)
	if err := e.SetData(cloudevents.ApplicationJSON, map[string]any{
		"value": map[string]any{
			"name": "projects/test-vickrey/databases/(default)/documents/drafts/0776cbd3-2e23-4911-99c1-f065a7cb052e",
		},
		"oldValue": map[string]any{},
	}); err != nil {
		t.Fatalf("set data: %v", err)
	}

	if err := HandleDraft(ctx, e); err != nil {
		t.Fatalf("HandleDraft: %v", err)
	}
}
