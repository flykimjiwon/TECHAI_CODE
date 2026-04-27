const { contextBridge, ipcRenderer } = require('electron')

// Expose menu events and native dialogs to React frontend
contextBridge.exposeInMainWorld('electronAPI', {
  onMenuEvent: (callback) => ipcRenderer.on('menu:openfolder', callback),
  onMenuSave: (callback) => ipcRenderer.on('menu:save', callback),
  onMenuSettings: (callback) => ipcRenderer.on('menu:settings', callback),
  onMenuPanel: (callback) => ipcRenderer.on('menu:panel', (_, panel) => callback(panel)),
  onMenuTerminal: (callback) => ipcRenderer.on('menu:terminal', callback),
  onMenuQuickOpen: (callback) => ipcRenderer.on('menu:quickopen', callback),
  onMenuAbout: (callback) => ipcRenderer.on('menu:about', callback),
  // Native folder dialog result
  onFolderOpened: (callback) => ipcRenderer.on('folder:opened', (_, dir) => callback(dir)),
})
