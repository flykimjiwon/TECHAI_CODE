// TECHAI IDE — Wails runtime wrapper with SSE fallback for Electron/browser mode

var _sseCallbacks = {};
var _sseCancelId = 0;
var _sseConnected = false;

function ensureSSE() {
  if (_sseConnected) return;
  _sseConnected = true;
  function connect() {
    try {
      var es = new EventSource('/api/events');
      var sseEvents = ['chat:chunk','chat:stream_start','chat:stream_done','chat:tool_start','chat:tool_done','chat:error','chat:cleared','file:changed','tree:refresh','term:output','preview:open'];
      var jsonEvents = { 'chat:tool_start': true, 'chat:tool_done': true };
      var noArgEvents = { 'chat:stream_start': true, 'chat:stream_done': true, 'chat:cleared': true, 'tree:refresh': true };
      sseEvents.forEach(function(name) {
        es.addEventListener(name, function(e) {
          try {
            (_sseCallbacks[name]||[]).forEach(function(x) {
              if (noArgEvents[name]) x.cb();
              else if (jsonEvents[name]) x.cb(JSON.parse(e.data));
              else x.cb(e.data);
            });
          } catch(err) {}
        });
      });
      es.onerror = function() { es.close(); setTimeout(connect, 2000); };
    } catch(err) { setTimeout(connect, 2000); }
  }
  connect();
}

function rt() {
  try { return window.runtime && window.runtime.EventsOnMultiple ? window.runtime : null; } catch(e) { return null; }
}

export function EventsOnMultiple(eventName, callback, maxCallbacks) {
  var r = rt();
  if (r) return r.EventsOnMultiple(eventName, callback, maxCallbacks);
  // SSE fallback
  ensureSSE();
  if (!_sseCallbacks[eventName]) _sseCallbacks[eventName] = [];
  var entry = { cb: callback, max: maxCallbacks, count: 0, id: ++_sseCancelId };
  _sseCallbacks[eventName].push(entry);
  return function() { _sseCallbacks[eventName] = (_sseCallbacks[eventName]||[]).filter(function(e){return e.id!==entry.id}) };
}

export function EventsOn(eventName, callback) {
  return EventsOnMultiple(eventName, callback, -1);
}

export function EventsOff(eventName) {
  var r = rt();
  if (r) return r.EventsOff(eventName);
  delete _sseCallbacks[eventName];
}

export function EventsOffAll() {
  var r = rt();
  if (r) return r.EventsOffAll();
  _sseCallbacks = {};
}

export function EventsOnce(eventName, callback) {
  return EventsOnMultiple(eventName, callback, 1);
}

export function EventsEmit(eventName) {
  var r = rt();
  if (r) return r.EventsEmit.apply(null, arguments);
  var data = arguments.length > 1 ? arguments[1] : undefined;
  fetch('/api/emit', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({name:eventName,data:data}) }).catch(function(){});
}

// Stubs for Wails-only functions (no-op in browser mode)
export function LogPrint(m) { var r = rt(); if (r) r.LogPrint(m); }
export function LogTrace(m) { var r = rt(); if (r) r.LogTrace(m); }
export function LogDebug(m) { var r = rt(); if (r) r.LogDebug(m); }
export function LogInfo(m) { var r = rt(); if (r) r.LogInfo(m); }
export function LogWarning(m) { var r = rt(); if (r) r.LogWarning(m); }
export function LogError(m) { var r = rt(); if (r) r.LogError(m); }
export function LogFatal(m) { var r = rt(); if (r) r.LogFatal(m); }
export function WindowReload() { var r = rt(); if (r) r.WindowReload(); }
export function WindowReloadApp() { var r = rt(); if (r) r.WindowReloadApp(); }
export function WindowSetAlwaysOnTop(b) { var r = rt(); if (r) r.WindowSetAlwaysOnTop(b); }
export function WindowSetSystemDefaultTheme() { var r = rt(); if (r) r.WindowSetSystemDefaultTheme(); }
export function WindowSetLightTheme() { var r = rt(); if (r) r.WindowSetLightTheme(); }
export function WindowSetDarkTheme() { var r = rt(); if (r) r.WindowSetDarkTheme(); }
export function WindowCenter() { var r = rt(); if (r) r.WindowCenter(); }
export function WindowSetTitle(t) { var r = rt(); if (r) r.WindowSetTitle(t); }
export function WindowFullscreen() { var r = rt(); if (r) r.WindowFullscreen(); }
export function WindowUnfullscreen() { var r = rt(); if (r) r.WindowUnfullscreen(); }
export function WindowIsFullscreen() { var r = rt(); if (r) return r.WindowIsFullscreen(); return false; }
export function WindowGetSize() { var r = rt(); if (r) return r.WindowGetSize(); return {w:1440,h:900}; }
export function WindowSetSize(w,h) { var r = rt(); if (r) r.WindowSetSize(w,h); }
export function WindowSetMaxSize(w,h) { var r = rt(); if (r) r.WindowSetMaxSize(w,h); }
export function WindowSetMinSize(w,h) { var r = rt(); if (r) r.WindowSetMinSize(w,h); }
export function WindowSetPosition(x,y) { var r = rt(); if (r) r.WindowSetPosition(x,y); }
export function WindowGetPosition() { var r = rt(); if (r) return r.WindowGetPosition(); return {x:0,y:0}; }
export function WindowHide() { var r = rt(); if (r) r.WindowHide(); }
export function WindowShow() { var r = rt(); if (r) r.WindowShow(); }
export function WindowMaximise() { var r = rt(); if (r) r.WindowMaximise(); }
export function WindowToggleMaximise() { var r = rt(); if (r) r.WindowToggleMaximise(); }
export function WindowUnmaximise() { var r = rt(); if (r) r.WindowUnmaximise(); }
export function WindowIsMaximised() { var r = rt(); if (r) return r.WindowIsMaximised(); return false; }
export function WindowMinimise() { var r = rt(); if (r) r.WindowMinimise(); }
export function WindowUnminimise() { var r = rt(); if (r) r.WindowUnminimise(); }
export function WindowSetBackgroundColour(R,G,B,A) { var r = rt(); if (r) r.WindowSetBackgroundColour(R,G,B,A); }
export function ScreenGetAll() { var r = rt(); if (r) return r.ScreenGetAll(); return []; }
export function WindowIsMinimised() { var r = rt(); if (r) return r.WindowIsMinimised(); return false; }
export function WindowIsNormal() { var r = rt(); if (r) return r.WindowIsNormal(); return true; }
export function BrowserOpenURL(url) { var r = rt(); if (r) r.BrowserOpenURL(url); else window.open(url,'_blank'); }
export function Environment() { var r = rt(); if (r) return r.Environment(); return {}; }
export function Quit() { var r = rt(); if (r) r.Quit(); }
export function Hide() { var r = rt(); if (r) r.Hide(); }
export function Show() { var r = rt(); if (r) r.Show(); }
export function ClipboardGetText() { var r = rt(); if (r) return r.ClipboardGetText(); return navigator.clipboard.readText(); }
export function ClipboardSetText(t) { var r = rt(); if (r) return r.ClipboardSetText(t); return navigator.clipboard.writeText(t); }
export function OnFileDrop(cb, u) { var r = rt(); if (r) return r.OnFileDrop(cb, u); }
export function OnFileDropOff() { var r = rt(); if (r) return r.OnFileDropOff(); }
export function CanResolveFilePaths() { var r = rt(); if (r) return r.CanResolveFilePaths(); return false; }
export function ResolveFilePaths(f) { var r = rt(); if (r) return r.ResolveFilePaths(f); return Promise.resolve(f); }
