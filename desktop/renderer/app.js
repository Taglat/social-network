// Desktop messenger renderer. Talks to the same backend as the website over
// HTTP + WebSocket, persists the login session (Electron cookie store), shows
// realtime online/offline presence, notifies on new messages, supports an
// interactive search with operators, and degrades gracefully when offline.

const EMOJIS = [
  "😀","😁","😂","🤣","😊","😍","😘","😎","🤔","😴","😢","😭","😡","👍","👎","👏",
  "🙏","💪","🔥","🎉","❤️","💔","⭐","✅","❌","💯","🥳","😅","🤗","😇","🤩","😋",
];

const state = {
  apiBase: localStorage.getItem("apiBase") || "http://localhost:8080",
  me: null,
  contacts: [],
  online: new Set(),
  active: null, // active contact id
  conversations: {}, // contactId -> messages[]
  unread: {}, // contactId -> count
  ws: null,
  connected: false, // websocket connected — authoritative "online" signal
};

const $ = (id) => document.getElementById(id);

// ---------- HTTP ----------
function wsBase() {
  return state.apiBase.replace(/^http/, "ws");
}
// Tell the main process which host may receive the session cookie. Must be
// awaited before the first request to that backend.
async function useBackend(apiBase) {
  state.apiBase = apiBase;
  localStorage.setItem("apiBase", apiBase);
  if (window.desktop) await window.desktop.setBackend(apiBase);
}
async function api(path, options = {}) {
  const res = await fetch(state.apiBase + path, {
    credentials: "include",
    ...options,
  });
  if (!res.ok) {
    let msg = res.statusText;
    try {
      const b = await res.json();
      if (b.error) msg = b.error;
    } catch {}
    const e = new Error(msg);
    e.status = res.status;
    throw e;
  }
  const txt = await res.text();
  return txt ? JSON.parse(txt) : null;
}

// ---------- Local cache (so messages are viewable offline) ----------
function cacheKey(suffix) {
  return `cache_${state.me ? state.me.id : "anon"}_${suffix}`;
}
function saveCache() {
  if (!state.me) return;
  localStorage.setItem(cacheKey("me"), JSON.stringify(state.me));
  localStorage.setItem(cacheKey("contacts"), JSON.stringify(state.contacts));
  localStorage.setItem(cacheKey("conversations"), JSON.stringify(state.conversations));
  localStorage.setItem("lastUserId", state.me.id);
}
function loadCache(userId) {
  const me = localStorage.getItem(`cache_${userId}_me`);
  if (me) state.me = JSON.parse(me);
  const c = localStorage.getItem(`cache_${userId}_contacts`);
  if (c) state.contacts = JSON.parse(c);
  const conv = localStorage.getItem(`cache_${userId}_conversations`);
  if (conv) state.conversations = JSON.parse(conv);
}

// ---------- Connectivity ----------
// "Online" is driven by whether the websocket is actually connected, not by the
// unreliable navigator.onLine flag. The OS online/offline events are only used
// as hints to retry the connection sooner.
function updateConnStatus() {
  const appVisible = !$("app-view").classList.contains("hidden");
  $("offline-banner").classList.toggle("hidden", !appVisible || state.connected);
  const el = $("conn-status");
  el.textContent = state.connected ? "● Online" : "● Offline";
  el.className = "status " + (state.connected ? "online" : "offline");
}
// The OS "back online" hint just makes the next retry happen immediately
// instead of waiting out the current backoff.
window.addEventListener("online", () => {
  if (state.me && !state.connected) connectWS();
});
window.addEventListener("offline", updateConnStatus);

// ---------- WebSocket ----------
let reconnectTimer = null;

function connectWS() {
  // One connection attempt at a time: onclose and the OS "online" event can
  // both ask to reconnect, and without this they stack up parallel sockets.
  clearTimeout(reconnectTimer);
  reconnectTimer = null;
  if (state.ws && state.ws.readyState <= 1) return; // CONNECTING or OPEN

  try {
    const ws = new WebSocket(wsBase() + "/ws");
    state.ws = ws;
    ws.onopen = () => {
      state.connected = true;
      updateConnStatus();
      resync(); // pull anything missed while the connection was down
    };
    ws.onclose = () => {
      if (state.ws !== ws) return; // superseded by a newer socket
      state.connected = false;
      updateConnStatus();
      // Retry while we are still logged in (handles transient drops).
      if (state.me && !reconnectTimer) reconnectTimer = setTimeout(connectWS, 2000);
    };
    ws.onmessage = (ev) => handleWsMessage(ev);
  } catch {
    state.connected = false;
    updateConnStatus();
    if (state.me && !reconnectTimer) reconnectTimer = setTimeout(connectWS, 2000);
  }
}

// After a drop, the contact list and the open conversation can both be stale:
// messages sent while we were offline were never pushed to us.
async function resync() {
  if (!state.me) return;
  try {
    state.contacts = (await api("/api/chat/partners")) || [];
  } catch {
    return; // still unreachable — the next reconnect will try again
  }
  if (state.active) {
    try {
      state.conversations[state.active] = (await api(`/api/messages/${state.active}`)) || [];
    } catch {
      /* keep the cached conversation */
    }
  }
  saveCache();
  renderContacts();
  renderMessages();
}

function handleWsMessage(ev) {
  let data;
  try {
    data = JSON.parse(ev.data);
  } catch {
    return;
  }
  if (data.type === "online_users") {
    state.online = new Set(data.userIds || []);
    renderContacts();
  } else if (data.type === "presence") {
    if (data.online) state.online.add(data.userId);
    else state.online.delete(data.userId);
    renderContacts();
  } else if (data.type === "message") {
    receiveMessage(data.message);
  }
}

function otherPartyOf(msg) {
  return msg.senderId === state.me.id ? msg.receiverId : msg.senderId;
}

function receiveMessage(msg) {
  if (msg.groupId) return; // desktop messenger handles private chats
  const other = otherPartyOf(msg);
  if (!state.conversations[other]) state.conversations[other] = [];
  // Avoid duplicates (the sender gets an echo of their own message).
  if (!state.conversations[other].some((m) => m.id === msg.id)) {
    state.conversations[other].push(msg);
  }
  saveCache();

  if (msg.senderId !== state.me.id) {
    if (state.active !== other || document.hidden) {
      state.unread[other] = (state.unread[other] || 0) + 1;
      notify(msg.senderName, msg.content);
    }
  }
  renderContacts();
  if (state.active === other) renderMessages();
}

function notify(title, body) {
  if (window.desktop) window.desktop.notify(title, body);
}

// ---------- Rendering ----------
function initials(name) {
  return name.split(" ").map((p) => p[0]).slice(0, 2).join("").toUpperCase();
}

function renderContacts() {
  const list = $("contacts-list");
  list.innerHTML = "";
  if (state.contacts.length === 0) {
    list.innerHTML = '<div class="empty" style="padding:14px">No contacts yet. Follow people on the website.</div>';
    return;
  }
  state.contacts.forEach((c) => {
    const name = `${c.firstName} ${c.lastName}`;
    const on = state.online.has(c.id);
    const unread = state.unread[c.id] || 0;
    const div = document.createElement("div");
    div.className = "contact" + (state.active === c.id ? " active" : "");
    div.innerHTML = `
      <div class="avatar">${initials(name)}</div>
      <div class="meta">
        <div class="name">${escapeHtml(name)}</div>
        <div class="muted">${on ? "Online" : "Offline"}</div>
      </div>
      ${unread ? `<span class="badge">${unread}</span>` : ""}
      <span class="dot ${on ? "on" : "off"}"></span>`;
    div.onclick = () => openConversation(c.id);
    list.appendChild(div);
  });
}

async function openConversation(contactId) {
  state.active = contactId;
  state.unread[contactId] = 0;
  const contact = state.contacts.find((c) => c.id === contactId);
  $("conv-title").textContent = contact
    ? `${contact.firstName} ${contact.lastName}`
    : "Conversation";

  // Try to load fresh history; if the request fails (offline) keep the cache.
  try {
    const msgs = await api(`/api/messages/${contactId}`);
    state.conversations[contactId] = msgs || [];
    saveCache();
  } catch {
    /* keep cached messages so they remain viewable offline */
  }
  renderContacts();
  renderMessages();
}

function renderMessages() {
  const box = $("messages");
  const all = state.conversations[state.active] || [];
  const predicate = parseSearch($("search").value);
  const msgs = all.filter((m) => predicate(m.content));

  box.innerHTML = "";
  if (!state.active) {
    box.innerHTML = '<div class="empty">Select a contact to start chatting.</div>';
    return;
  }
  if (msgs.length === 0) {
    box.innerHTML = `<div class="empty">${
      $("search").value ? "No messages match your search." : "No messages yet."
    }</div>`;
    return;
  }
  msgs.forEach((m) => {
    const mine = m.senderId === state.me.id;
    const div = document.createElement("div");
    div.className = "bubble" + (mine ? " mine" : "");
    div.innerHTML = `${mine ? "" : `<div class="sender">${escapeHtml(m.senderName)}</div>`}
      ${escapeHtml(m.content)}
      <div class="time">${new Date(m.createdAt).toLocaleTimeString()}</div>`;
    box.appendChild(div);
  });
  box.scrollTop = box.scrollHeight;
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, (c) =>
    ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c])
  );
}

// ---------- Search with operators ----------
// Supports: plain term (include, substring), -term/!term (exclude),
// ~term (fuzzy subsequence), and numeric comparators >N <N >=N <=N =N !=N.
function parseSearch(query) {
  const q = query.trim().toLowerCase();
  if (!q) return () => true;
  const tokens = q.split(/\s+/);
  const numRe = /^(>=|<=|!=|>|<|=)(\d+(?:\.\d+)?)$/;

  return (contentRaw) => {
    const content = contentRaw.toLowerCase();
    const numbers = (content.match(/-?\d+(?:\.\d+)?/g) || []).map(Number);
    return tokens.every((t) => {
      const num = t.match(numRe);
      if (num) {
        const op = num[1];
        const val = parseFloat(num[2]);
        return numbers.some((n) => compare(n, op, val));
      }
      if (t.startsWith("-") || t.startsWith("!")) {
        return !content.includes(t.slice(1));
      }
      if (t.startsWith("~")) {
        return fuzzy(content, t.slice(1));
      }
      return content.includes(t);
    });
  };
}
function compare(n, op, v) {
  switch (op) {
    case ">": return n > v;
    case "<": return n < v;
    case ">=": return n >= v;
    case "<=": return n <= v;
    case "=": return n === v;
    case "!=": return n !== v;
  }
  return false;
}
// fuzzy: true if all characters of term appear in order within content.
function fuzzy(content, term) {
  let i = 0;
  for (const ch of content) {
    if (ch === term[i]) i++;
    if (i === term.length) return true;
  }
  return term.length === 0;
}

// ---------- Send ----------
function sendMessage() {
  const input = $("msg-input");
  const content = input.value.trim();
  if (!content || !state.active) return;
  // Offline guard: blocked while the realtime connection is down.
  if (!state.connected) {
    alert("You are offline. You cannot send messages until the connection is restored.");
    return;
  }
  state.ws.send(JSON.stringify({ type: "private_message", to: state.active, content }));
  input.value = "";
  $("emoji-pop").classList.add("hidden");
}

// ---------- Auth flow ----------
async function showApp() {
  $("login-view").classList.add("hidden");
  $("app-view").classList.remove("hidden");
  $("me-name").textContent = `${state.me.firstName} ${state.me.lastName}`;

  // Always attempt to refresh contacts and open the realtime connection; fall
  // back to cached data when the requests fail (offline).
  try {
    state.contacts = (await api("/api/chat/partners")) || [];
    saveCache();
  } catch {
    /* use cached contacts */
  }
  connectWS();
  updateConnStatus();
  renderContacts();
  renderMessages();
}

function showLogin() {
  $("app-view").classList.add("hidden");
  $("login-view").classList.remove("hidden");
  updateConnStatus(); // hides the offline banner on the login screen
}

async function doLogin() {
  $("login-error").textContent = "";
  await useBackend($("api-base").value.trim() || "http://localhost:8080");
  const email = $("email").value.trim();
  const password = $("password").value;
  try {
    state.me = await api("/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    saveCache();
    await showApp();
  } catch (e) {
    $("login-error").textContent =
      e.status === 401 ? "Invalid email or password" : e.message;
  }
}

async function doLogout() {
  try {
    await api("/api/logout", { method: "POST" });
  } catch {
    /* the server may be unreachable — the local session is dropped regardless */
  }
  // Drop the persisted token explicitly. A failed logout request sends no
  // Set-Cookie, so without this the app would silently log back in on restart.
  if (window.desktop) await window.desktop.clearSession();
  clearTimeout(reconnectTimer);
  reconnectTimer = null;
  if (state.ws) state.ws.close();
  // Clear cached session so reopening the app asks for login again.
  if (state.me) localStorage.removeItem(`cache_${state.me.id}_me`);
  localStorage.removeItem("lastUserId");
  state.me = null;
  state.contacts = [];
  state.conversations = {};
  state.active = null;
  state.unread = {};
  showLogin();
}

// ---------- Init ----------
function buildEmojiPicker() {
  const pop = $("emoji-pop");
  EMOJIS.forEach((e) => {
    const b = document.createElement("button");
    b.type = "button";
    b.textContent = e;
    b.onclick = () => {
      $("msg-input").value += e;
      $("msg-input").focus();
    };
    pop.appendChild(b);
  });
}

function wireEvents() {
  $("login-btn").onclick = doLogin;
  $("password").addEventListener("keydown", (e) => e.key === "Enter" && doLogin());
  $("logout-btn").onclick = doLogout;
  $("send-btn").onclick = sendMessage;
  $("msg-input").addEventListener("keydown", (e) => e.key === "Enter" && sendMessage());
  $("search").addEventListener("input", renderMessages);
  $("emoji-btn").onclick = () => $("emoji-pop").classList.toggle("hidden");
  $("register-link").onclick = (e) => {
    e.preventDefault();
    const base = $("api-base").value.trim() || "http://localhost:8080";
    // Frontend website runs on port 3000 in the default setup.
    const site = base.replace(/:8080$/, ":3000");
    if (window.desktop) window.desktop.openExternal(site + "/register");
  };
}

async function init() {
  $("api-base").value = state.apiBase;
  buildEmojiPicker();
  wireEvents();
  await useBackend(state.apiBase);

  // Try to resume an existing session (the token is persisted by the main process).
  try {
    state.me = await api("/api/me");
    saveCache();
    await showApp();
  } catch (e) {
    if (e.status === 401) {
      // A valid response: simply not logged in.
      showLogin();
    } else {
      // Network error -> offline. Load the last cached session for read-only
      // viewing if there is one, otherwise show the login screen.
      const last = localStorage.getItem("lastUserId");
      if (last) loadCache(last);
      if (state.me) await showApp();
      else showLogin();
    }
  }
}

init();
