# Social Network — Desktop Messenger (Electron)

A cross-platform (Windows / Linux / macOS) desktop messenger for the social
network. It connects to the **same backend** as the website, so a website user
and a desktop user can chat in realtime.

## Features

- Email + password login (sessions persist across restarts via Electron's
  cookie store — you stay logged in until logout or session expiry).
- "Register" link opens the social network website in the browser.
- Realtime private chat over WebSocket (interoperates with the website).
- Online/offline presence: contacts' status updates live; when a user logs out
  or loses connection all their followers see them go offline immediately.
- Native notifications when a new message arrives.
- Emoji picker.
- Offline mode: a banner warns when the connection drops; cached messages stay
  readable, but sending is blocked with a warning.
- Interactive search (results update as you type) with operators:
  - `term` include (substring), `-term` / `!term` exclude, `~term` fuzzy
  - numeric comparators over numbers in messages: `>5`, `<10`, `>=3`, `<=7`, `=4`, `!=2`

## Run (development)

Make sure the backend is running (e.g. `docker compose up` from the repo root),
then:

```bash
cd desktop
npm install
npm start
```

On the login screen the **Backend server** field defaults to
`http://localhost:8080`; change it if your backend runs elsewhere.

## Build installers

```bash
npm run dist:win     # Windows NSIS installer
npm run dist:linux   # Linux AppImage
npm run dist:mac     # macOS dmg
```

Output is written to `desktop/dist/`. The Linux target is an **AppImage**; the
Windows target is an NSIS installer suitable for installing inside a VM.
