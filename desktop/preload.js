// Preload bridge: exposes a minimal, safe API to the renderer.
const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("desktop", {
  openExternal: (url) => ipcRenderer.invoke("open-external", url),
  notify: (title, body) => ipcRenderer.invoke("notify", { title, body }),
  setBackend: (apiBase) => ipcRenderer.invoke("set-backend", apiBase),
  clearSession: () => ipcRenderer.invoke("clear-session"),
});
