package draft

import "time"

type Draft struct {
	LeagueID  string   `firestore:"league_id" json:"league_id"`
	CreatedDT time.Time `firestore:"created_dt" json:"created_dt"`
	Users     []User   `firestore:"users" json:"users"`
	Players   []Player `firestore:"players" json:"players"`
}

type User struct {
	DisplayName   string        `firestore:"display_name" json:"display_name"`
	CurrentBudget int           `firestore:"current_budget" json:"current_budget"`
	Roster        []interface{} `firestore:"roster" json:"roster"`
}

type Player struct {
	FullName         string   `firestore:"full_name" json:"full_name"`
	SearchRank       int      `firestore:"search_rank" json:"search_rank"`
	InjuryStatus     string   `firestore:"injury_status" json:"injury_status"`
	Status           string   `firestore:"status" json:"status"`
	FantasyPositions []string `firestore:"fantasy_positions" json:"fantasy_positions"`
}
