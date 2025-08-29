import { doc, onSnapshot } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-firestore.js';

const mainDocRef = doc(db, 'drafts', '2c64afce-43f6-4c8f-ac0b-263d903e2e18', 'pages', userName);
const localRoot = document.getElementById('root');

onSnapshot(mainDocRef, (snap) => {
  const html = snap.data()?.html ?? '<h1>No HTML found</h1>';
  const remoteRoot = new DOMParser().parseFromString(html, 'text/html').getElementById('root');
  if (localRoot && remoteRoot) localRoot.innerHTML = remoteRoot.innerHTML;
});
