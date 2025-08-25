package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// cors sets the minimum headers needed for simple CORS and pre-flight handling.
func cors(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return false
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "test-vickrey"
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("firestore.NewClient: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	docRef := client.Collection("teams").Doc(creds.Username)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.Error(w, "team not found", http.StatusNotFound)
			return
		}
		log.Printf("firestore get: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := docSnap.Data()
	hashVal, hasHash := data["passwordHash"].(string)

	switch {
	case !hasHash:
		newHash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("password hash: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_, err = docRef.Update(ctx, []firestore.Update{
			{Path: "passwordHash", Value: string(newHash)},
			{Path: "updatedAt", Value: time.Now()},
		})
		if err != nil {
			log.Printf("firestore update: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	case bcrypt.CompareHashAndPassword([]byte(hashVal), []byte(creds.Password)) != nil:
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		log.Printf("firebase.NewApp: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Printf("firebase Auth: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := authClient.CustomToken(ctx, creds.Username)
	if err != nil {
		log.Printf("custom token: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func init() {
	functions.HTTP("Handler", Handler)
}
