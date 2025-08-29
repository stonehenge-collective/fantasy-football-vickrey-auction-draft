import { doc, onSnapshot } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-firestore.js';
import { signInWithCustomToken } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-auth.js';

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
