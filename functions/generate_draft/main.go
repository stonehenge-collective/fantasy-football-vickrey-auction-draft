package main

/*
   Creates an initial state document for a fantasy-football auction draft.

   * Pulls league members from the Sleeper API (filters out bots).
   * Pulls the full NFL player list from the Sleeper API, caches the raw
     response to players_raw.json, and keeps only draft-eligible players.
   * Builds a JSON “state” object:
       {
         "league_id": …,
         "created_dt": …,
         "users": [ … ],
         "players": [ … ]   // ordered by search_rank ascending
       }
   * Persists that state:
       • to Firestore under drafts/{generated-uuid}
       • to game_state.json on disk
     and writes the filtered players slice to draft_players.json.

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
	"sort"
	"time"

	cloudfirestore "cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

const (
	defaultLeagueID  = "1263349443621568512"
	defaultProjectID = "test-vickrey"

	playersURL   = "https://api.sleeper.app/v1/players/nfl"
	usersURLTmpl = "https://api.sleeper.app/v1/league/%s/users"

	rawPlayersFile   = "players_raw.json"
	draftPlayersFile = "draft_players.json"
	gameStateFile    = "game_state.json"

	budget = 100
)

type sleeperUser struct {
	DisplayName string `json:"display_name"`
	IsBot       bool   `json:"is_bot"`
}

type userState struct {
	DisplayName   string   `json:"display_name" firestore:"display_name"`
	CurrentBudget int      `json:"current_budget" firestore:"current_budget"`
	Roster        []string `json:"roster" firestore:"roster"`
}

type playerState struct {
	FullName         string   `json:"full_name" firestore:"full_name"`
	SearchRank       int      `json:"search_rank" firestore:"search_rank"`
	InjuryStatus     string   `json:"injury_status" firestore:"injury_status"`
	Status           string   `json:"status" firestore:"status"`
	FantasyPositions []string `json:"fantasy_positions" firestore:"fantasy_positions"`
}

type draftState struct {
	LeagueID string        `json:"league_id" firestore:"league_id"`
	Created  time.Time     `json:"created_dt" firestore:"created_dt"`
	Users    []userState   `json:"users" firestore:"users"`
	Players  []playerState `json:"players" firestore:"players"`
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
	if err != nil {
		var raw []byte
		raw, players, err = fetchPlayers(ctx)
		if err != nil {
			log.Fatalf("fetch players: %v", err)
		}
		if err := os.WriteFile(rawPlayersFile, raw, 0o644); err != nil {
			log.Fatalf("write raw players file: %v", err)
		}
	}

	filtered := filterPlayers(players)
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].SearchRank < filtered[j].SearchRank })

	if err := writeJSONFile(draftPlayersFile, filtered); err != nil {
		log.Fatalf("write draft players file: %v", err)
	}

	state := draftState{
		LeagueID: *leagueID,
		Created:  time.Now(),
		Users:    users,
		Players:  filtered,
	}

	if err := writeJSONFile(gameStateFile, state); err != nil {
		log.Fatalf("write game state file: %v", err)
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
			DisplayName:   u.DisplayName,
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

func filterPlayers(all map[string]playerState) []playerState {
	allowed := map[string]struct{}{
		"QB": {}, "RB": {}, "WR": {}, "TE": {},
	}
	out := make([]playerState, 0, len(all))
	for _, p := range all {
		if p.SearchRank >= 250 || p.SearchRank == 0 {
			continue
		}
		if p.FullName == "Player Invalid" {
			continue
		}
		for _, pos := range p.FantasyPositions {
			if _, ok := allowed[pos]; ok {
				out = append(out, p)
				break
			}
		}
	}
	return out
}

func writeJSONFile(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
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
