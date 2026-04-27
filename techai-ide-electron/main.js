const { app, BrowserWindow, Menu, shell } = require('electron')
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
async function startGoServer() {
  serverPort = await findPort(8080)

  const serverPath = getServerPath()
  if (!fs.existsSync(serverPath)) {
    console.error('Go server not found:', serverPath)
    return false
  }

  // Make executable on Unix
  if (process.platform !== 'win32') {
    fs.chmodSync(serverPath, '755')
  }

  const cwd = process.argv[2] || process.cwd()

  goServer = spawn(serverPath, ['--port', String(serverPort), '--cwd', cwd], {
    stdio: ['pipe', 'pipe', 'pipe'],
    env: { ...process.env, TECHAI_PORT: String(serverPort) },
  })

  goServer.stdout.on('data', (data) => console.log('[GO]', data.toString().trim()))
  goServer.stderr.on('data', (data) => console.error('[GO-ERR]', data.toString().trim()))
  goServer.on('exit', (code) => console.log('[GO] exited with code', code))

  // Wait for server to be ready
  await new Promise((resolve) => {
    const check = () => {
      const req = net.createConnection({ port: serverPort }, () => {
        req.destroy()
        resolve()
      })
      req.on('error', () => setTimeout(check, 100))
    }
    setTimeout(check, 200)
  })

  return true
}

function createMenu() {
  const isMac = process.platform === 'darwin'

  const template = [
    ...(isMac ? [{ role: 'appMenu' }] : []),
    {
      label: 'File',
      submenu: [
        { label: 'Open Folder...', accelerator: 'CmdOrCtrl+O', click: () => mainWindow?.webContents.send('menu:openfolder') },
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
  mainWindow = new BrowserWindow({
    width: 1440,
    height: 900,
    minWidth: 1024,
    minHeight: 640,
    title: 'TECHAI IDE',
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

  mainWindow.on('closed', () => { mainWindow = null })

  // Open external links in browser
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url)
    return { action: 'deny' }
  })
}

app.whenReady().then(async () => {
  console.log('Starting TECHAI IDE...')

  const started = await startGoServer()
  if (!started) {
    console.error('Failed to start Go server')
    app.quit()
    return
  }

  console.log(`Go server running on port ${serverPort}`)
  createMenu()
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
