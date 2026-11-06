// Electron main process. Creates the window and persists the login session.
//
// Session persistence: the session token is captured from the backend's
// Set-Cookie header and stored on disk by this process (see below for why the
// cookie store itself cannot be relied on). The app stays logged in until the
// session expires or the user logs out.
const { app, BrowserWindow, shell, ipcMain, Notification, session, safeStorage } = require("electron");
const path = require("path");
const fs = require("fs");

let mainWindow;

// The backend authenticates with a SameSite=Lax session cookie. From the
// renderer's file:// origin every backend request is cross-site, so Chromium
// neither reliably stores nor sends that cookie. To keep the desktop client
// authenticated WITHOUT changing the backend, the main process captures the
// session token from the login response and re-attaches it to every subsequent
// API/WebSocket request itself. The token is persisted to disk so the user
// stays logged in across app restarts (until logout or session expiry).
let sessionToken = null;

// Host (hostname:port) of the backend. The token is only ever captured from and
// sent to this host, so the session cookie can never leak to a third party the
// app happens to contact. The renderer updates it from the "Backend server"
// field before it issues any request.
let backendHost = "localhost:8080";

const tokenFile = () => path.join(app.getPath("userData"), "session.token");

// The token is stored encrypted with the OS keychain/credential store when the
// platform provides one, falling back to plain text where it does not (some
// bare Linux desktops). The "v1:" prefix marks an encrypted payload.
function loadToken() {
  try {
    const raw = fs.readFileSync(tokenFile());
    const text = raw.toString("utf8");
    if (text.startsWith("v1:")) {
      sessionToken = safeStorage.decryptString(raw.subarray(3)) || null;
    } else {
      sessionToken = text.trim() || null;
    }
  } catch {
    sessionToken = null;
  }
}
function saveToken() {
  try {
    if (!sessionToken) {
      fs.rmSync(tokenFile(), { force: true });
      return;
    }
    if (safeStorage.isEncryptionAvailable()) {
      const enc = safeStorage.encryptString(sessionToken);
      fs.writeFileSync(tokenFile(), Buffer.concat([Buffer.from("v1:"), enc]));
    } else {
      fs.writeFileSync(tokenFile(), sessionToken);
    }
  } catch {
    /* ignore persistence errors */
  }
}

// Is this request going to the backend we are authenticated against?
function isBackendRequest(url) {
  try {
    const u = new URL(url);
    return (
      ["http:", "https:", "ws:", "wss:"].includes(u.protocol) &&
      u.host === backendHost
    );
  } catch {
    return false;
  }
}

function installAuthForwarding() {
  const ses = session.defaultSession;

  // Capture / clear the session token from backend Set-Cookie responses.
  ses.webRequest.onHeadersReceived((details, callback) => {
    const headers = details.responseHeaders || {};
    const key = isBackendRequest(details.url)
      ? Object.keys(headers).find((k) => k.toLowerCase() === "set-cookie")
      : undefined;
    if (key) {
      for (const raw of headers[key]) {
        const m = /(?:^|;\s*)session=([^;]*)/.exec(raw);
        if (m) {
          const val = m[1];
          // An empty value (logout) clears the session.
          sessionToken = val && val !== '""' ? val : null;
          saveToken();
        }
      }
    }
    callback({ responseHeaders: headers });
  });

  // Attach the captured token to outgoing backend requests (HTTP + WebSocket).
  ses.webRequest.onBeforeSendHeaders((details, callback) => {
    if (sessionToken && isBackendRequest(details.url)) {
      details.requestHeaders["Cookie"] = `session=${sessionToken}`;
    }
    callback({ requestHeaders: details.requestHeaders });
  });
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1000,
    height: 720,
    minWidth: 720,
    minHeight: 520,
    title: "Social Network Messenger",
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
      // The renderer talks to the backend API on another origin (localhost:8080)
      // with credentials. Relaxing webSecurity avoids CORS friction for this
      // first-party desktop client; cookies are still stored per backend host.
      webSecurity: false,
    },
  });

  mainWindow.loadFile(path.join(__dirname, "renderer", "index.html"));
  mainWindow.setMenuBarVisibility(false);
}

// Open external links (e.g. the "register on the website" link) in the browser.
ipcMain.handle("open-external", async (_e, url) => {
  await shell.openExternal(url);
});

// Point the auth forwarding at the backend the user selected on the login
// screen. The renderer awaits this before it issues its first request.
ipcMain.handle("set-backend", (_e, apiBase) => {
  try {
    backendHost = new URL(apiBase).host;
  } catch {
    /* keep the previous host on a malformed value */
  }
});

// Drop the stored session. Called on logout so the app asks for credentials on
// the next start even when the logout request itself never reached the server
// (e.g. the user logged out while offline, so no Set-Cookie cleared the token).
ipcMain.handle("clear-session", () => {
  sessionToken = null;
  saveToken();
});

// Native desktop notification when a message arrives.
ipcMain.handle("notify", (_e, { title, body }) => {
  if (Notification.isSupported()) {
    new Notification({ title, body }).show();
  }
});

app.whenReady().then(() => {
  loadToken();
  installAuthForwarding();
  createWindow();
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") app.quit();
});
