package function

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
)

func TestHandleDraft(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:9090")
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-vickrey")

	ctx := context.Background()
	ctx = context.WithValue(ctx, "ce-type", "google.cloud.firestore.document.v1.updated")

	_, _ = firestore.NewClient(ctx, "test-vickrey") // ensures emulator is reachable

	event := FirestoreEvent{
		Value: struct {
			Name   string                 `json:"name"`
			Fields map[string]any         `json:"fields"`
		}{
			Name: "projects/test-vickrey/databases/(default)/documents/drafts/0776cbd3-2e23-4911-99c1-f065a7cb052e",
		},
	}

	if err := HandleDraft(ctx, event); err != nil {
		t.Fatalf("HandleDraft: %v", err)
	}
}
