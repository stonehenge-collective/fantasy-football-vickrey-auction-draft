package http

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request, ", r.Method)
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "test-vickrey"
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
	docRef := client.Collection("messages").NewDoc()
	_, err = docRef.Set(ctx, map[string]interface{}{
		"content":   "Hello Firestore",
		"timestamp": time.Now(),
	})
	if err != nil {
		log.Printf("firestore write: %v", err)
		http.Error(w, "failed to write to Firestore", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Wrote document %s to Firestore in project %s\n", docRef.ID, projectID)
}

func init() {
	functions.HTTP("Handler", Handler)
}