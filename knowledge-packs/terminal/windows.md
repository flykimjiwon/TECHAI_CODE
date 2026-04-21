# Windows PowerShell & CMD Reference

## File Operations

### List Files (ls)
```powershell
# PowerShell
Get-ChildItem
Get-ChildItem -Recurse  # 재귀적
ls *.txt                # 필터

# CMD
dir
dir /s  # 재귀적
```

### Read Files (cat)
```powershell
# PowerShell
Get-Content file.txt
cat file.txt            # 별칭

# CMD
type file.txt
```

### Write Files
```powershell
# PowerShell
Set-Content file.txt "content"
"content" | Out-File file.txt

# CMD
echo content > file.txt
```

## Network

### IP Address (ipconfig)
```powershell
# PowerShell
Get-NetIPAddress
Get-NetIPConfiguration

# CMD
ipconfig
ipconfig /all
```

### Test Connection (ping)
```powershell
# PowerShell
Test-NetConnection google.com
Test-NetConnection -Port 80

# CMD
ping google.com
```

### HTTP Request (curl)
```powershell
# PowerShell
Invoke-WebRequest https://api.example.com/data
Invoke-RestMethod https://api.example.com/data  # JSON 자동 파싱

# CMD (curl.exe 필요)
curl https://api.example.com/data
```

## Process Management

### List Processes (ps)
```powershell
# PowerShell
Get-Process
Get-Process chrome

# CMD
tasklist
tasklist | findstr chrome
```

### Kill Process
```powershell
# PowerShell
Stop-Process -Name chrome
Stop-Process -Id 1234

# CMD
taskkill /IM chrome.exe /F
taskkill /PID 1234 /F
```

### Start Process
```powershell
# PowerShell
Start-Process notepad.exe
Start-Process "C:\Program Files\app.exe" -ArgumentList "--flag"

# CMD
start notepad.exe
```

## Environment Variables
```powershell
# PowerShell
$env:PATH
$env:USERNAME
$env:PATH += ";C:\new\path"

# CMD
echo %PATH%
echo %USERNAME%
set MY_VAR=value
```

## Package Management (winget)
```powershell
winget search nodejs
winget install Microsoft.PowerShell
winget upgrade --all
winget list
```

## 자주 쓰는 명령어

### Find Files
```powershell
# PowerShell
Get-ChildItem -Recurse -Filter *.log
Get-ChildItem -Recurse | Where-Object { $_.Length -gt 1MB }

# CMD
dir /s *.log
```

### Services
```powershell
# PowerShell
Get-Service
Start-Service wuauserv
Stop-Service wuauserv
Restart-Service wuauserv

# CMD
sc query
sc start wuauserv
sc stop wuauserv
```

### System Info
```powershell
# PowerShell
Get-ComputerInfo
systeminfo

# CMD
systeminfo
```

### Clipboard
```powershell
# PowerShell
Get-Content file.txt | Set-Clipboard
Get-Clipboard
```
