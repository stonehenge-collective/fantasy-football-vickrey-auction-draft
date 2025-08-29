import { initializeApp } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-app.js';
import { getAnalytics } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-analytics.js';
import { getFirestore } from 'https://www.gstatic.com/firebasejs/12.1.0/firebase-firestore.js';
import {
  getAuth,
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
