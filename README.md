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
firebase init emulators (then !space! -> enter to select firestore emulator. 9090 for emulator port, 4000 for ui)
firebase emulators:start --only firestore

gcloud beta emulators firestore start --host-port=localhost:9090

(bash, from functions/registration)FUNCTION_TARGET=Handler LOCAL_ONLY=true FIRESTORE_EMULATOR_HOST=localhost:9090 go run cmd/main.go

Hit http://localhost:8080 and watch documents appear in the emulator UI (http://localhost:4000).
