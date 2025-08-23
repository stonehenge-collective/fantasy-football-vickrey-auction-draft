package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
)

func main() {
	ctx := context.Background()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "test-vickrey"
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
	defer client.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		docRef := client.Collection("messages").NewDoc()
		_, err := docRef.Set(ctx, map[string]interface{}{
			"content":   "Hello Firestore",
			"timestamp": time.Now(),
		})
		if err != nil {
			http.Error(w, "failed to write to Firestore", http.StatusInternalServerError)
			log.Printf("firestore write: %v", err)
			return
		}

		fmt.Fprintf(w, "Wrote document %s to Firestore in project %s\n", docRef.ID, projectID)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
