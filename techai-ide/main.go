package main

import (
	"embed"
	"fmt"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Debug log to file (helps diagnose Windows issues)
	logFile, _ := os.Create("techai-ide-debug.log")
	if logFile != nil {
		defer logFile.Close()
		fmt.Fprintf(logFile, "TECHAI IDE starting...\n")
		fmt.Fprintf(logFile, "OS: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Fprintf(logFile, "CWD: %s\n", func() string { d, _ := os.Getwd(); return d }())
	}
	log := func(msg string) {
		if logFile != nil {
			fmt.Fprintf(logFile, "%s\n", msg)
		}
	}

	// Check embedded assets
	if entries, err := assets.ReadDir("frontend/dist"); err != nil {
		log("ERROR reading embedded assets: " + err.Error())
	} else {
		log(fmt.Sprintf("Embedded assets: %d entries in frontend/dist", len(entries)))
		for _, e := range entries {
			log("  " + e.Name())
		}
	}

	log("Creating app...")
	app := NewApp()
	log("Building menu...")
	appMenu := buildMenu(app)

	log("Starting Wails...")
	err := wails.Run(&options.App{
		Menu: appMenu,
		Title:            "TECHAI CODE IDE",
		Width:            1440,
		Height:           900,
		MinWidth:         1024,
		MinHeight:        640,
		DisableResize:    false,
		Frameless:        false,
		StartHidden:      false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 12, A: 255}, // #0a0a0c
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		EnableDefaultContextMenu: true,
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		log("ERROR: " + err.Error())
		println("Error:", err.Error())
	}
	log("App exited.")
}

func buildMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	// File menu
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open Folder...", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:openfolder")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Save", keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:save")
	})
	fileMenu.AddText("Save All", keys.Combo("s", keys.CmdOrCtrlKey, keys.ShiftKey), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:saveall")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Close Tab", keys.CmdOrCtrl("w"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:closetab")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Settings...", keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:settings")
	})

	// Edit menu — Role-based for native Undo/Redo/Cut/Copy/Paste/SelectAll
	appMenu.Append(menu.EditMenu())

	// View menu
	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Explorer", keys.CmdOrCtrl("1"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:panel", "files")
	})
	viewMenu.AddText("Search", keys.CmdOrCtrl("2"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:panel", "search")
	})
	viewMenu.AddText("Git", keys.CmdOrCtrl("3"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:panel", "git")
	})
	viewMenu.AddSeparator()
	viewMenu.AddText("Toggle Sidebar", keys.CmdOrCtrl("b"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:panel", "files")
	})
	viewMenu.AddText("Toggle Terminal", keys.CmdOrCtrl("j"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:terminal")
	})
	viewMenu.AddSeparator()
	viewMenu.AddText("Quick Open...", keys.CmdOrCtrl("p"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:quickopen")
	})
	viewMenu.AddText("Theme...", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:theme")
	})

	// Terminal menu
	termMenu := appMenu.AddSubmenu("Terminal")
	termMenu.AddText("New Terminal", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:newterminal")
	})

	// Help menu
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("About TECHAI IDE", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:about")
	})
	helpMenu.AddText("Keyboard Shortcuts", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:shortcuts")
	})

	return appMenu
}
