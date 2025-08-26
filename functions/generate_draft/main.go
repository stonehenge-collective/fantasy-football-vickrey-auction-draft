package main

/*
   Creates an initial state document for a fantasy-football auction draft.
   * Pulls league members from the Sleeper API (filters out bots).
   * Pulls the full NFL player list from the Sleeper API, keeps only
     the fields needed for drafting, and also saves the raw response
     to disk so it can be reused without calling the API again.
   * Builds a JSON “state” object:
       {
         "league_id": …,
         "created_dt": …,
         "users": [ {display_name, current_budget, roster}, … ],
         "players": {player_id: {...}, … }
       }
   * Persists that state as a document in Firestore under
     the collection “drafts/{generated-uuid}”.

   Run locally, e.g.:
       go run main.go --league 1263349443621568512 --project my-gcp-project
*/

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	cloudfirestore "cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

const (
	defaultLeagueID = "1263349443621568512"
	defaultProjectID = "test-vickrey"
	playersURL      = "https://api.sleeper.app/v1/players/nfl"
	usersURLTmpl    = "https://api.sleeper.app/v1/league/%s/users"
	rawPlayersFile  = "players_raw.json"
	budget          = 100
)

type sleeperUser struct {
	DisplayName string `json:"display_name"`
	IsBot       bool   `json:"is_bot"`
}

type userState struct {
	DisplayName  string   `json:"display_name" firestore:"display_name"`
	CurrentBudget int      `json:"current_budget" firestore:"current_budget"`
	Roster        []string `json:"roster" firestore:"roster"`
}

type playerState struct {
	FullName        string   `json:"full_name" firestore:"full_name"`
	SearchRank       int      `json:"search_rank" firestore:"search_rank"`
	InjuryStatus     string   `json:"injury_status" firestore:"injury_status"`
	Status           string   `json:"status" firestore:"status"`
	FantasyPositions []string `json:"fantasy_positions" firestore:"fantasy_positions"`
}

type draftState struct {
	LeagueID string                     `json:"league_id" firestore:"league_id"`
	Created  time.Time                  `json:"created_dt" firestore:"created_dt"`
	Users    []userState                `json:"users" firestore:"users"`
	Players  map[string]playerState     `json:"players" firestore:"players"`
}

func main() {
	leagueID := flag.String("league", defaultLeagueID, "Sleeper league ID")
	projectID := flag.String("project", defaultProjectID, "GCP project ID (defaults to $GOOGLE_CLOUD_PROJECT)")
	flag.Parse()

	if *projectID == "" {
		*projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if *projectID == "" {
		log.Fatalf("GCP project ID must be supplied via --project or $GOOGLE_CLOUD_PROJECT")
	}

	ctx := context.Background()

	users, err := fetchLeagueUsers(ctx, *leagueID)
	if err != nil {
		log.Fatalf("fetch league users: %v", err)
	}

	players, err := loadPlayersFromFile()
	if err != nil {                       // cache miss ➜ call API and save cache
		var rawPlayers []byte
		rawPlayers, players, err = fetchPlayers(ctx)
		if err != nil {
			log.Fatalf("fetch players: %v", err)
		}
		if err := os.WriteFile(rawPlayersFile, rawPlayers, 0o644); err != nil {
			log.Fatalf("write raw players file: %v", err)
		}
	}

	state := draftState{
		LeagueID: *leagueID,
		Created:  time.Now(),
		Users:    users,
		Players:  players,
	}

	docID, err := saveDraft(ctx, *projectID, state)
	if err != nil {
		log.Fatalf("save draft: %v", err)
	}

	fmt.Printf("Draft state saved to Firestore at drafts/%s\n", docID)
}

func loadPlayersFromFile() (map[string]playerState, error) {
	raw, err := os.ReadFile(rawPlayersFile)
	if err != nil {
		return nil, err
	}
	var players map[string]playerState
	if err := json.Unmarshal(raw, &players); err != nil {
		return nil, err
	}
	return players, nil
}

func fetchLeagueUsers(ctx context.Context, leagueID string) ([]userState, error) {
	url := fmt.Sprintf(usersURLTmpl, leagueID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get users: %w", err)
	}
	defer resp.Body.Close()

	var apiUsers []sleeperUser
	if err := json.NewDecoder(resp.Body).Decode(&apiUsers); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}

	var result []userState
	for _, u := range apiUsers {
		if u.IsBot {
			continue
		}
		result = append(result, userState{
			DisplayName:  u.DisplayName,
			CurrentBudget: budget,
			Roster:        []string{},
		})
	}
	return result, nil
}

func fetchPlayers(ctx context.Context) ([]byte, map[string]playerState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, playersURL, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http get players: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read players body: %w", err)
	}

	var api map[string]playerState
	if err := json.Unmarshal(raw, &api); err != nil {
		return nil, nil, fmt.Errorf("decode players: %w", err)
	}
	return raw, api, nil
}

func saveDraft(ctx context.Context, projectID string, state draftState) (string, error) {
	client, err := cloudfirestore.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("firestore client: %w", err)
	}
	defer client.Close()

	docID := uuid.NewString()
	if _, err := client.Collection("drafts").Doc(docID).Set(ctx, state); err != nil {
		return "", fmt.Errorf("firestore set: %w", err)
	}
	return docID, nil
}
