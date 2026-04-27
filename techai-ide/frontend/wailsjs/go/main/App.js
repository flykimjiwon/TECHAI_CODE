// Copyright 2025-2026 Kim Jiwon (flykimjiwon). All rights reserved.
// TECHAI CODE — github.com/flykimjiwon/TECHAI_CODE
// Wails API wrapper with REST API fallback for Electron/browser mode
// If window.go.main.App exists (Wails), use it. Otherwise, use REST API.

function api(method, args) {
  return fetch('/api/' + method, {
    method: 'POST',
    headers: {'Content-Type':'application/json'},
    body: args !== undefined ? JSON.stringify(args) : undefined
  }).then(function(r) { return r.text().then(function(t) { try{return JSON.parse(t)}catch(e){return t} }) });
}

function get(method, params) {
  var qs = params ? '?' + new URLSearchParams(params).toString() : '';
  return fetch('/api/' + method + qs).then(function(r) { return r.text().then(function(t) { try{return JSON.parse(t)}catch(e){return t} }) });
}

function wails() {
  try { return window['go']['main']['App']; } catch(e) { return null; }
}

function call(name, args) {
  var w = wails();
  if (w && w[name]) return w[name].apply(null, args);
  return null;
}

export function ClearChat() {
  return call('ClearChat', []) || api('clearChat');
}

export function DeleteFile(arg1) {
  return call('DeleteFile', [arg1]) || api('deleteFile', {path: arg1});
}

export function DeleteSession(arg1) {
  return call('DeleteSession', [arg1]) || api('deleteSession', {id: arg1});
}

export function ExportChat() {
  return call('ExportChat', []) || api('exportChat');
}

export function FileExists(arg1) {
  return call('FileExists', [arg1]) || get('fileExists', {path: arg1});
}

export function GetAvailableShells() {
  return call('GetAvailableShells', []) || get('getShells');
}

export function GetChatHistory() {
  return call('GetChatHistory', []) || get('chatHistory');
}

export function GetCurrentShell() {
  return call('GetCurrentShell', []) || get('getCurrentShell');
}

export function GetCwd() {
  return call('GetCwd', []) || get('getCwd');
}

export function GetFileIcon(arg1) {
  return call('GetFileIcon', [arg1]) || Promise.resolve('file');
}

export function GetGitBranches() {
  return call('GetGitBranches', []) || get('gitBranches');
}

export function GetGitGraph(arg1) {
  return call('GetGitGraph', [arg1]) || get('gitGraph');
}

export function GetGitInfo() {
  return call('GetGitInfo', []) || get('gitInfo');
}

export function GetKnowledgePacks() {
  return call('GetKnowledgePacks', []) || get('knowledgePacks');
}

export function GetModel() {
  return call('GetModel', []) || get('getModel');
}

export function GetRecentProjects() {
  return call('GetRecentProjects', []) || get('recentProjects');
}

export function GetSearchLimit() {
  return call('GetSearchLimit', []) || Promise.resolve(100);
}

export function GetSettings() {
  return call('GetSettings', []) || get('getSettings');
}

export function GitCheckout(arg1) {
  return call('GitCheckout', [arg1]) || api('gitCheckout', {branch: arg1});
}

export function GitCommit(arg1) {
  return call('GitCommit', [arg1]) || api('gitCommit', {message: arg1});
}

export function GitCreateBranch(arg1) {
  return call('GitCreateBranch', [arg1]) || api('gitCreateBranch', {name: arg1});
}

export function GitDiff() {
  return call('GitDiff', []) || get('gitDiff');
}

export function GitDiffFile(arg1) {
  return call('GitDiffFile', [arg1]) || get('gitDiffFile', {path: arg1});
}

export function GitLog(arg1) {
  return call('GitLog', [arg1]) || get('gitLog');
}

export function GitPull() {
  return call('GitPull', []) || api('gitPull');
}

export function GitPush() {
  return call('GitPush', []) || api('gitPush');
}

export function GitStage(arg1) {
  return call('GitStage', [arg1]) || api('gitStage', {path: arg1});
}

export function GitUnstage(arg1) {
  return call('GitUnstage', [arg1]) || api('gitUnstage', {path: arg1});
}

export function ListFiles(arg1, arg2) {
  return call('ListFiles', [arg1, arg2]) || get('listFiles', {path: arg1 || '.'});
}

export function ListSessions() {
  return call('ListSessions', []) || get('listSessions');
}

export function LoadSession(arg1) {
  return call('LoadSession', [arg1]) || api('loadSession', {id: arg1});
}

export function OpenFolder() {
  var w = wails();
  if (w && w['OpenFolder']) return w['OpenFolder']();
  return new Promise(function(resolve) {
    if (window.electronAPI && window.electronAPI.onFolderOpened) {
      var done = false;
      window.electronAPI.onFolderOpened(function(dir) { if (!done) { done = true; resolve(dir); } });
      setTimeout(function() { if (!done) { done = true; resolve(''); } }, 60000);
    } else { resolve(''); }
  });
}

export function OpenInBrowser(arg1) {
  var w = wails();
  if (w && w['OpenInBrowser']) return w['OpenInBrowser'](arg1);
  window.open(arg1, '_blank');
  return Promise.resolve();
}

export function ReadFile(arg1) {
  return call('ReadFile', [arg1]) || get('readFile', {path: arg1});
}

export function RenameFile(arg1, arg2) {
  return call('RenameFile', [arg1, arg2]) || api('renameFile', {oldPath: arg1, newPath: arg2});
}

export function ResizeTerminal(arg1, arg2) {
  return call('ResizeTerminal', [arg1, arg2]) || Promise.resolve();
}

export function SaveDroppedFile(arg1, arg2) {
  return call('SaveDroppedFile', [arg1, arg2]) || api('writeFile', {path: arg1, content: arg2});
}

export function SaveSession(arg1) {
  return call('SaveSession', [arg1]) || api('saveSession', {title: arg1});
}

export function SaveSettings(arg1, arg2, arg3) {
  return call('SaveSettings', [arg1, arg2, arg3]) || api('saveSettings', {baseURL: arg1, apiKey: arg2, model: arg3});
}

export function SearchInFiles(arg1, arg2) {
  return call('SearchInFiles', [arg1, arg2]) || get('search', {pattern: arg1, path: arg2 || '.'});
}

export function SendMessage(arg1) {
  var w = wails();
  if (w && w['SendMessage']) return w['SendMessage'](arg1);
  fetch('/api/sendMessage', {method:'POST', headers:{'Content-Type':'text/plain'}, body: arg1});
  return Promise.resolve();
}

export function SetCwd(arg1) {
  return call('SetCwd', [arg1]) || api('setCwd', {path: arg1});
}

export function SetLanguage(arg1) {
  return call('SetLanguage', [arg1]) || api('setLanguage', {lang: arg1});
}

export function SetShell(arg1) {
  return call('SetShell', [arg1]) || api('setShell', {shell: arg1});
}

export function StartLiveServer(arg1) {
  return call('StartLiveServer', [arg1]) || api('startLiveServer', {dir: arg1});
}

export function StartTerminal() {
  return call('StartTerminal', []) || api('startTerminal');
}

export function StopTerminal() {
  return call('StopTerminal', []) || api('stopTerminal');
}

export function ToggleKnowledgePack(arg1, arg2) {
  return call('ToggleKnowledgePack', [arg1, arg2]) || api('toggleKnowledgePack', {id: arg1, enabled: arg2});
}

export function WalkProject() {
  return call('WalkProject', []) || get('walkProject');
}

export function WriteFile(arg1, arg2) {
  return call('WriteFile', [arg1, arg2]) || api('writeFile', {path: arg1, content: arg2});
}

export function WriteTerminal(arg1) {
  var w = wails();
  if (w && w['WriteTerminal']) return w['WriteTerminal'](arg1);
  return api('writeTerminal', arg1);
}
