gcloud functions deploy html_builder \
 --gen2 \
 --runtime=go125 \
 --entry-point=HandleDraft \
 --region=us-central1 \
--trigger-event-filters="type=google.cloud.firestore.document.v1.written" \
 --trigger-event-filters="database=(default)" \
 --trigger-event-filters-path-pattern="document=drafts/{draftId}" \
 --trigger-location=nam5 \
 --min-instances=0
