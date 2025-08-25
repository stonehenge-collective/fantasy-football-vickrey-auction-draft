# fantasy-football-vickrey-auction-draft

https://stonehenge-collective.github.io/fantasy-football-vickrey-auction-draft/

go mod tidy

local dev dependencies:
https://go.dev/dl/
https://cloud.google.com/sdk/docs/install
https://www.docker.com/products/docker-desktop/
https://buildpacks.io/docs/for-platform-operators/how-to/integrate-ci/pack/
https://www.java.com/en/download/manual.jsp
https://www.oracle.com/java/technologies/downloads/

gcloud init
gcloud auth login

npm install -g firebase-tools
firebase login
firebase init emulators (then !space! -> enter to select firestore and auth emulators. leave port 9099 for auth, 9090 for emulator port, 4000 for ui)
firebase emulators:start

Open your Firebase project in https://console.firebase.google.com
In the left-hand nav, click ⚙️ Project settings ▸ Service accounts.
Under Firebase Admin SDK, pick Go (language is irrelevant; it just influences the snippet).
Press Generate new private key → Generate.
A file named like project-id-xxxx.json downloads immediately—this is your serviceAccount.json.
Store it somewhere outside your repo (e.g. ~/secrets/serviceAccount.json).

(bash, from functions/registration)FUNCTION_TARGET=Handler LOCAL_ONLY=true FIRESTORE_EMULATOR_HOST=localhost:9090 FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 GOOGLE_APPLICATION_CREDENTIALS="C:\Users\johnm\repos\fantasy-football-vickrey-auction-draft\test-vickrey-firebase-adminsdk-fbsvc-57a145d978.json" go run cmd/main.go

Hit http://localhost:8080 and watch documents appear in the emulator UI (http://localhost:4000).
curl -X POST http://localhost:8080 -d '{"username":"johnor","password":"s3cr3t"}'
