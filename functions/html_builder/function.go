package function

import (
	"context"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/firestore"
)

// FirestoreEvent mirrors the minimal shape of the CloudEvent payload we need.
type FirestoreEvent struct {
	OldValue struct{} `json:"oldValue"`
	Value    struct {
		Name   string                 `json:"name"`
		Fields map[string]any         `json:"fields"` // not used – we re-read the doc
	} `json:"value"`
}

// Draft mirrors the Firestore document.
type Draft struct {
	LeagueID string   `firestore:"league_id"`
	Users    []User   `firestore:"users"`
	Players  []Player `firestore:"players"`
}

type User struct {
	DisplayName   string        `firestore:"display_name"`
	CurrentBudget int           `firestore:"current_budget"`
	Roster        []interface{} `firestore:"roster"`
}

type Player struct {
	FullName         string   `firestore:"full_name"`
	SearchRank       int      `firestore:"search_rank"`
	InjuryStatus     string   `firestore:"injury_status"`
	Status           string   `firestore:"status"`
	FantasyPositions []string `firestore:"fantasy_positions"`
}

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
      import { initializeApp } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-app.js';
      import { getAnalytics } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-analytics.js';
      import {
        getFirestore,
        doc,
        onSnapshot,
      } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-firestore.js';
      import {
        getAuth,
        signInWithCustomToken,
        indexedDBLocalPersistence,
        setPersistence,
      } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-auth.js';

      const firebaseConfig = {
        apiKey: 'AIzaSyBF7ONgQ0LCYGcf2pRpcUSjH4eaKDSNkwE',
        authDomain: 'test-vickrey.firebaseapp.com',
        projectId: 'test-vickrey',
        storageBucket: 'test-vickrey.firebasestorage.app',
        messagingSenderId: '373721638486',
        appId: '1:373721638486:web:1884aeaf06132f1047fdf7',
        measurementId: 'G-LGCLMMK0EJ',
      };

      const app = initializeApp(firebaseConfig);
      getAnalytics(app);
      const auth = getAuth(app);
      setPersistence(auth, indexedDBLocalPersistence);

      const db = getFirestore(app);
      const mainDocRef = doc(db, 'drafts', '%s', 'pages', 'public');
      const localRoot = document.getElementById('root');

      onSnapshot(mainDocRef, (snap) => {
        const html = snap.data()?.html ?? '<h1>No HTML found</h1>';
        const remoteRoot = new DOMParser()
          .parseFromString(html, 'text/html')
          .getElementById('root');
        if (localRoot && remoteRoot) localRoot.innerHTML = remoteRoot.innerHTML;
      });

      const form = document.getElementById('joinDraftForm');
      form.addEventListener('submit', joinDraft);

      async function joinDraft(event) {
        event.preventDefault();

        const username = event.target.sleeperUsername.value.trim();
        const password = event.target.draftPassword.value;

        try {
          const res = await fetch('https://vickrey-registration-373721638486.us-east1.run.app/', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password }),
          });

          if (res.status === 404) {
            alert('Team name not found');
          } else if (res.status === 401) {
            alert('Invalid password');
          } else if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Request failed with status ' + res.status);
          } else {
            const token_response = await res.json();
            await signInWithCustomToken(auth, token_response.token);
            const teamPageRef = doc(db, 'drafts', '%s', 'pages', username);

            onSnapshot(teamPageRef, (snap) => {
              const data = snap.data();
              const html = data?.html ?? '<h1>No HTML found</h1>';
              document.open();
              document.write(html);
              document.close();
            });
          }
        } catch (err) {
          console.error(err);
          alert('There was a problem joining the draft. See console for details.');
        }
      }
    </script>
  </body>
</html>
`

// HandleDraft regenerates the public page every time a draft document is created or updated.
func HandleDraft(ctx context.Context, e FirestoreEvent) error {
	if ctx.Value("ce-type") == "google.cloud.firestore.document.v1.deleted" {
		return nil
	}

	re := regexp.MustCompile(`^projects\/[^\/]+\/databases\/\(default\)\/documents\/(.+)$`)
	docPath := re.ReplaceAllString(e.Value.Name, `$1`)
	if !strings.HasPrefix(docPath, "drafts/") {
		return fmt.Errorf("unexpected path %q", docPath)
	}
	parts := strings.Split(docPath, "/")
	draftID := parts[len(parts)-1]

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
    if projectID == "" {
        return fmt.Errorf("GOOGLE_CLOUD_PROJECT not set")
    }
    client, err := firestore.NewClient(ctx, projectID)
    if err != nil {
        return fmt.Errorf("firestore client: %w", err)
    }
    defer client.Close()

	snap, err := client.Doc(docPath).Get(ctx)
	if err != nil {
		return fmt.Errorf("read draft: %w", err)
	}

	var draft Draft
	if err := snap.DataTo(&draft); err != nil {
		return fmt.Errorf("decode draft: %w", err)
	}

	// Build comma-separated, HTML-escaped lists
	userNames := make([]string, len(draft.Users))
	for i, u := range draft.Users {
		userNames[i] = html.EscapeString(u.DisplayName)
	}
	playerNames := make([]string, len(draft.Players))
	for i, p := range draft.Players {
		playerNames[i] = html.EscapeString(p.FullName)
	}

	root := fmt.Sprintf(
		"<p>Users: %s</p><p>Players: %s</p>",
		strings.Join(userNames, ", "),
		strings.Join(playerNames, ", "),
	)

	htmlPage := fmt.Sprintf(pageTemplate, root, draftID, draftID)

	if _, err := client.Collection("drafts").
		Doc(draftID).
		Collection("pages").
		Doc("public").
		Set(ctx, map[string]any{"html": htmlPage}); err != nil {
		return fmt.Errorf("write public page: %w", err)
	}

	return nil
}
