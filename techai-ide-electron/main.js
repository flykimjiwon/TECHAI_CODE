const { app, BrowserWindow, Menu, shell, dialog } = require('electron')
const { spawn } = require('child_process')
const path = require('path')
const fs = require('fs')
const net = require('net')

let mainWindow = null
let goServer = null
let serverPort = 8080

// Find the Go server binary
function getServerPath() {
  const isPackaged = app.isPackaged
  const platform = process.platform === 'win32' ? 'techai-server.exe' : 'techai-server'

  if (isPackaged) {
    return path.join(process.resourcesPath, 'go-server', platform)
  }
  return path.join(__dirname, 'go-server', platform)
}

// Find an available port
function findPort(start) {
  return new Promise((resolve) => {
    const server = net.createServer()
    server.listen(start, () => {
      const port = server.address().port
      server.close(() => resolve(port))
    })
    server.on('error', () => resolve(findPort(start + 1)))
  })
}

// Start Go backend server
let goServerError = ''

async function startGoServer() {
  serverPort = await findPort(8080)

  const serverPath = getServerPath()
  console.log('[BOOT] Server path:', serverPath)
  console.log('[BOOT] Exists:', fs.existsSync(serverPath))
  console.log('[BOOT] Platform:', process.platform, 'Arch:', process.arch)
  console.log('[BOOT] Packaged:', app.isPackaged)
  console.log('[BOOT] ResourcesPath:', process.resourcesPath)

  if (!fs.existsSync(serverPath)) {
    goServerError = 'Go server not found: ' + serverPath
    console.error(goServerError)
    return false
  }

  // Make executable on Unix
  if (process.platform !== 'win32') {
    fs.chmodSync(serverPath, '755')
  }

  const cwd = process.argv[2] || process.cwd()
  console.log('[BOOT] CWD:', cwd, 'Port:', serverPort)

  goServer = spawn(serverPath, ['--port', String(serverPort), '--cwd', cwd], {
    stdio: ['pipe', 'pipe', 'pipe'],
    env: { ...process.env, TECHAI_PORT: String(serverPort) },
    windowsHide: true,
  })

  let serverDied = false
  goServer.stdout.on('data', (data) => console.log('[GO]', data.toString().trim()))
  goServer.stderr.on('data', (data) => {
    const msg = data.toString().trim()
    console.error('[GO-ERR]', msg)
    goServerError += msg + '\n'
  })
  goServer.on('exit', (code) => {
    console.log('[GO] exited with code', code)
    serverDied = true
    if (!goServerError) goServerError = 'Go server exited with code ' + code
  })

  // Wait for server to be ready (with timeout)
  const ready = await new Promise((resolve) => {
    let attempts = 0
    const maxAttempts = 50 // 5 seconds max
    const check = () => {
      if (serverDied) { resolve(false); return }
      if (attempts++ >= maxAttempts) { resolve(false); return }
      const req = net.createConnection({ port: serverPort }, () => {
        req.destroy()
        resolve(true)
      })
      req.on('error', () => setTimeout(check, 100))
    }
    setTimeout(check, 200)
  })

  if (!ready) {
    if (!goServerError) goServerError = 'Go server failed to start (timeout)'
    console.error('[BOOT]', goServerError)
    return false
  }

  return true
}

function createMenu() {
  const isMac = process.platform === 'darwin'

  const template = [
    ...(isMac ? [{ role: 'appMenu' }] : []),
    {
      label: 'File',
      submenu: [
        { label: 'Open Folder...', accelerator: 'CmdOrCtrl+O', click: () => openFolderDialog() },
        { type: 'separator' },
        { label: 'Save', accelerator: 'CmdOrCtrl+S', click: () => mainWindow?.webContents.send('menu:save') },
        { type: 'separator' },
        { label: 'Settings', accelerator: 'CmdOrCtrl+,', click: () => mainWindow?.webContents.send('menu:settings') },
        { type: 'separator' },
        isMac ? { role: 'close' } : { role: 'quit' },
      ],
    },
    { role: 'editMenu' },
    {
      label: 'View',
      submenu: [
        { label: 'Explorer', accelerator: 'CmdOrCtrl+1', click: () => mainWindow?.webContents.send('menu:panel', 'files') },
        { label: 'Search', accelerator: 'CmdOrCtrl+2', click: () => mainWindow?.webContents.send('menu:panel', 'search') },
        { label: 'Git', accelerator: 'CmdOrCtrl+3', click: () => mainWindow?.webContents.send('menu:panel', 'git') },
        { type: 'separator' },
        { label: 'Toggle Terminal', accelerator: 'CmdOrCtrl+J', click: () => mainWindow?.webContents.send('menu:terminal') },
        { label: 'Quick Open', accelerator: 'CmdOrCtrl+P', click: () => mainWindow?.webContents.send('menu:quickopen') },
        { type: 'separator' },
        { role: 'toggleDevTools' },
        { role: 'togglefullscreen' },
      ],
    },
    {
      label: 'Terminal',
      submenu: [
        { label: 'New Terminal', click: () => mainWindow?.webContents.send('menu:newterminal') },
      ],
    },
    {
      label: 'Help',
      submenu: [
        { label: 'About TECHAI IDE', click: () => mainWindow?.webContents.send('menu:about') },
      ],
    },
  ]

  Menu.setApplicationMenu(Menu.buildFromTemplate(template))
}

function createWindow() {
  const iconPath = app.isPackaged
    ? path.join(process.resourcesPath, 'icons', process.platform === 'win32' ? 'icon.ico' : 'icon.png')
    : path.join(__dirname, 'icons', process.platform === 'win32' ? 'icon.ico' : 'icon.png')

  mainWindow = new BrowserWindow({
    width: 1440,
    height: 900,
    minWidth: 1024,
    minHeight: 640,
    title: 'TECHAI IDE',
    icon: iconPath,
    backgroundColor: '#0a0a0c',
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js'),
    },
    ...(process.platform === 'darwin' ? {
      titleBarStyle: 'hiddenInset',
      trafficLightPosition: { x: 12, y: 10 },
    } : {}),
  })

  // Load the React frontend via Go server
  mainWindow.loadURL(`http://localhost:${serverPort}`)

  // If Go server dies while running, show error
  if (goServer) {
    goServer.on('exit', (code) => {
      if (mainWindow && code !== 0 && code !== null) {
        mainWindow.loadURL('data:text/html;charset=utf-8,' + encodeURIComponent(
          `<!DOCTYPE html><html><head><style>body{margin:0;background:#0a0a0c;color:#e4e4e7;font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;flex-direction:column;gap:12px}h1{color:#f87171;font-size:18px}p{color:#888;font-size:13px}</style></head><body><h1>Backend server crashed (code ${code})</h1><p>Restart the application to try again.</p></body></html>`
        ))
      }
    })
  }

  mainWindow.on('closed', () => { mainWindow = null })

  // Open external links in browser
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url)
    return { action: 'deny' }
  })
}

// Error window when Go server fails
function createErrorWindow(errorMsg) {
  const win = new BrowserWindow({
    width: 600, height: 400, title: 'TECHAI IDE — Error',
    backgroundColor: '#0a0a0c',
    webPreferences: { nodeIntegration: false, contextIsolation: true },
  })
  const html = `<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
    body{margin:0;background:#0a0a0c;color:#e4e4e7;font-family:-apple-system,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;flex-direction:column;gap:16px;padding:40px}
    h1{color:#f87171;font-size:20px;margin:0} pre{background:#1a1a2e;padding:16px;border-radius:8px;font-size:12px;max-width:100%;overflow:auto;border:1px solid #333;white-space:pre-wrap;word-break:break-all}
    .info{font-size:12px;color:#888;text-align:center}
  </style></head><body>
    <h1>Go Server Failed to Start</h1>
    <pre>${errorMsg.replace(/</g,'&lt;').replace(/>/g,'&gt;')}</pre>
    <div class="info">
      Platform: ${process.platform} | Arch: ${process.arch}<br>
      Resources: ${process.resourcesPath}<br>
      Server path: ${getServerPath()}
    </div>
  </body></html>`
  win.loadURL('data:text/html;charset=utf-8,' + encodeURIComponent(html))
}

// Native folder open dialog
async function openFolderDialog() {
  if (!mainWindow) return
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openDirectory'],
    title: 'Open Project Folder',
  })
  if (result.canceled || result.filePaths.length === 0) return

  const dir = result.filePaths[0]

  // Tell Go server to switch cwd
  try {
    await fetch(`http://localhost:${serverPort}/api/setCwd`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path: dir }),
    })
  } catch (e) {
    console.error('Failed to set cwd:', e)
  }

  // Update window title
  mainWindow.setTitle(`TECHAI IDE — ${dir}`)

  // Notify frontend to refresh
  mainWindow.webContents.send('folder:opened', dir)
}

app.whenReady().then(async () => {
  console.log('Starting TECHAI IDE...')

  // Set Dock icon on macOS (dev mode)
  if (process.platform === 'darwin' && app.dock) {
    const devIcon = path.join(__dirname, 'icons', 'icon.png')
    if (fs.existsSync(devIcon)) {
      app.dock.setIcon(devIcon)
    }
  }

  const started = await startGoServer()
  createMenu()

  if (!started) {
    console.error('Failed to start Go server:', goServerError)
    // Show error window instead of quitting
    createErrorWindow(goServerError)
    return
  }

  console.log(`Go server running on port ${serverPort}`)
  createWindow()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  if (goServer) {
    goServer.kill()
    goServer = null
  }
  if (process.platform !== 'darwin') app.quit()
})

app.on('before-quit', () => {
  if (goServer) {
    goServer.kill()
    goServer = null
  }
})
