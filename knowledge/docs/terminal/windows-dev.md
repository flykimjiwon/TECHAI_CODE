# Windows 개발 환경 Reference

> Windows CMD / PowerShell / WSL2 환경에서 개발할 때 자주 쓰는 명령어와 설정 모음.
> 한국어 요청 → Windows 명령어 변환을 위해 택가이코드 LLM이 참조하는 문서.

---

## 드라이브 이동

```cmd
# CMD에서 드라이브 이동 — cd만으로는 안 됨, 드라이브문자: 단독 입력
D:
E:
cd D:\projects\myapp  # 드라이브 이동 + 경로 이동 동시 불가 (CMD)

# PowerShell에서는 cd/Set-Location으로 드라이브 이동 가능
Set-Location D:\projects\myapp
cd D:\projects\myapp  # PS에서는 이것도 동작함
```

---

## PowerShell vs CMD 차이

| 항목 | CMD | PowerShell |
|-----|-----|-----------|
| 기본 셸 | 구형 (Windows 초기부터) | 모던 (.NET 기반) |
| 스크립트 확장자 | `.bat`, `.cmd` | `.ps1` |
| 파이프 대상 | 텍스트 문자열 | 객체 (Object) |
| 파일 목록 | `dir` | `Get-ChildItem` (또는 `ls`, `dir`) |
| 파일 내용 | `type 파일명` | `Get-Content 파일명` (또는 `cat`) |
| 환경변수 읽기 | `%VAR%` | `$env:VAR` |
| 환경변수 설정 | `set VAR=값` (임시) | `$env:VAR = "값"` (임시) |
| 조건문 | `if`, `for` (제한적) | 완전한 스크립팅 지원 |
| 원격 관리 | 제한적 | WinRM 기반 풍부한 지원 |
| 권장 여부 | 레거시 호환 | 신규 스크립트는 PS 권장 |

```powershell
# 현재 PowerShell 버전 확인
$PSVersionTable.PSVersion

# PowerShell 7 (크로스 플랫폼) 설치 — winget 사용
winget install Microsoft.PowerShell
```

---

## Windows Terminal 설정

```json
// %LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_*\LocalState\settings.json
{
    "defaultProfile": "{PowerShell GUID}",
    "profiles": {
        "defaults": {
            "font": {
                "face": "CaskaydiaCove Nerd Font",  // 나눔고딕코딩, D2Coding도 가능
                "size": 13
            },
            "colorScheme": "One Half Dark"
        }
    }
}
```

```powershell
# Windows Terminal 설치 (미설치 시)
winget install Microsoft.WindowsTerminal
```

---

## WSL2 사용법

```powershell
# WSL2 설치 (Windows 10 2004 이상 / Windows 11)
wsl --install                        # Ubuntu 기본 설치
wsl --install -d Ubuntu-22.04        # 특정 배포판 지정

# 배포판 목록 확인
wsl --list --verbose                 # 줄여서 wsl -l -v
wsl --list --online                  # 설치 가능한 배포판 목록

# 기본 배포판 변경
wsl --set-default Ubuntu-22.04

# WSL 버전 변경
wsl --set-version Ubuntu-22.04 2     # WSL2로 업그레이드

# WSL 진입/종료
wsl                                  # 기본 배포판으로 진입
wsl -d Ubuntu-22.04                  # 특정 배포판으로 진입
exit                                 # WSL 셸 종료
wsl --shutdown                       # 모든 WSL 인스턴스 종료

# Windows 파일을 WSL에서 접근
cd /mnt/c/Users/username/projects    # C 드라이브 → /mnt/c/
cd /mnt/d/                           # D 드라이브 → /mnt/d/

# WSL에서 Windows 명령 실행
explorer.exe .                       # 현재 WSL 경로를 탐색기로 열기
notepad.exe 파일명                   # Windows 메모장으로 파일 열기
```

---

## 환경변수 영구 설정

```cmd
:: CMD에서 영구 설정 (현재 사용자)
setx MY_VAR "my_value"
setx PATH "%PATH%;C:\new\tool\bin"   :: ⚠️ 1024자 제한 있음

:: 시스템 전체 영구 설정 (관리자 권한 필요)
setx MY_VAR "my_value" /M
```

```powershell
# PowerShell에서 영구 설정
[System.Environment]::SetEnvironmentVariable("MY_VAR", "my_value", "User")    # 현재 사용자
[System.Environment]::SetEnvironmentVariable("MY_VAR", "my_value", "Machine") # 시스템 전체 (관리자)

# 설정 확인
[System.Environment]::GetEnvironmentVariable("MY_VAR", "User")

# PATH에 영구 추가
$oldPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
[System.Environment]::SetEnvironmentVariable("PATH", "$oldPath;C:\new\tool\bin", "User")

# GUI로 열기 (System Properties)
sysdm.cpl                            # 시스템 속성 → 고급 → 환경 변수
```

---

## Windows Defender 제외 설정 (개발 성능 향상)

```powershell
# 관리자 권한 PowerShell에서 실행
# 개발 폴더 전체를 실시간 검사에서 제외 → 빌드 속도 대폭 향상

Add-MpPreference -ExclusionPath "C:\projects"
Add-MpPreference -ExclusionPath "C:\Users\username\go"
Add-MpPreference -ExclusionPath "C:\Users\username\AppData\Roaming\npm"
Add-MpPreference -ExclusionPath "C:\Users\username\.cargo"

# 특정 프로세스 제외 (Go, Node 컴파일러 등)
Add-MpPreference -ExclusionProcess "go.exe"
Add-MpPreference -ExclusionProcess "node.exe"

# 제외 목록 확인
Get-MpPreference | Select-Object -ExpandProperty ExclusionPath
```

---

## winget 패키지 관리

```powershell
# 패키지 검색
winget search git
winget search "visual studio code"

# 패키지 설치
winget install Git.Git
winget install Microsoft.VisualStudioCode
winget install OpenJS.NodeJS.LTS
winget install GoLang.Go
winget install Python.Python.3
winget install Docker.DockerDesktop

# 설치된 패키지 목록
winget list

# 업그레이드
winget upgrade Git.Git           # 특정 패키지
winget upgrade --all             # 전체 업그레이드

# 패키지 제거
winget uninstall Git.Git

# 패키지 정보
winget show Git.Git
```

---

## 개발 도구 일괄 설치

```powershell
# 개발 환경 초기 셋업 스크립트 (관리자 PowerShell)
$packages = @(
    "Git.Git",
    "Microsoft.VisualStudioCode",
    "OpenJS.NodeJS.LTS",
    "GoLang.Go",
    "Python.Python.3",
    "Microsoft.WindowsTerminal",
    "Microsoft.PowerShell"
)

foreach ($pkg in $packages) {
    Write-Host "설치 중: $pkg"
    winget install $pkg --silent --accept-package-agreements
}
```

---

## 포트 확인 및 해제

```cmd
:: 특정 포트 사용 중인 프로세스 찾기
netstat -ano | findstr :8080
netstat -ano | findstr :3000

:: PID로 프로세스 이름 확인
tasklist | findstr 1234

:: 프로세스 강제 종료 (PID 사용)
taskkill /PID 1234 /F

:: 프로세스 이름으로 종료
taskkill /IM node.exe /F
taskkill /IM go.exe /F
```

```powershell
# PowerShell로 포트 점유 프로세스 한 번에 찾아서 종료
$port = 8080
$process = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
if ($process) {
    Stop-Process -Id $process.OwningProcess -Force
    Write-Host "포트 $port 점유 프로세스 종료 완료"
} else {
    Write-Host "포트 $port 를 사용하는 프로세스 없음"
}
```

---

## 긴 경로 이름 활성화

Windows는 기본적으로 경로 길이를 260자로 제한함. Node.js `node_modules` 등에서 문제 발생.

```powershell
# 방법 1: 레지스트리 편집 (관리자 PowerShell)
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem" `
    -Name "LongPathsEnabled" -Value 1 -Type DWord

# 방법 2: 그룹 정책 편집기 (gpedit.msc)
# 컴퓨터 구성 → 관리 템플릿 → 시스템 → 파일 시스템
# → "Win32 긴 경로 사용" → 사용으로 설정

# Git도 긴 경로 허용 설정
git config --system core.longpaths true
```

---

## 줄바꿈 문제 (CRLF/LF)

Windows는 `\r\n`, Linux/macOS는 `\n` 사용. Git 협업 시 충돌 원인.

```bash
# Windows에서 Git 체크아웃 시 CRLF로 변환, 커밋 시 LF로 변환 (권장)
git config --global core.autocrlf true

# 변환 없이 있는 그대로 사용 (WSL 환경이나 Linux 전용 프로젝트)
git config --global core.autocrlf false

# 체크아웃 시 변환 없음, 커밋 시 LF로만 변환 (Linux 서버 배포 프로젝트)
git config --global core.autocrlf input

# 프로젝트별 설정 (.gitattributes)
# * text=auto eol=lf    ← 모든 파일을 LF로 통일
```

```powershell
# 파일의 줄바꿈 형식 확인
(Get-Content 파일명 -Raw) -match "`r`n"  # True면 CRLF

# CRLF → LF 변환 (PowerShell)
(Get-Content 파일명 -Raw) -replace "`r`n", "`n" | Set-Content 파일명 -NoNewline
```

---

## 한국어 인코딩 (UTF-8)

```cmd
:: CMD 한국어 깨짐 해결 — UTF-8로 코드페이지 변경
chcp 65001

:: 현재 코드페이지 확인
chcp
:: 949 = EUC-KR (한국어 기본), 65001 = UTF-8, 437 = 영어 기본
```

```powershell
# PowerShell 기본 인코딩을 UTF-8로 설정
# $PROFILE 파일에 추가 (영구 설정)
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

# 파일 읽기/쓰기 인코딩 지정
Get-Content 파일명 -Encoding UTF8
Set-Content 파일명 -Encoding UTF8 -Value "내용"
```

---

## 유용한 Windows 전용 명령어

```powershell
# 시스템 정보
systeminfo                           # 상세 시스템 정보
Get-ComputerInfo                     # PowerShell 시스템 정보

# 디스크 사용량
Get-PSDrive C                        # C 드라이브 사용량
wmic logicaldisk get size,freespace,caption  # 전체 드라이브 용량

# 클립보드 활용
Get-Content 파일명 | clip            # 파일 내용을 클립보드에 복사
"텍스트" | clip                      # 텍스트를 클립보드에 복사

# 현재 위치를 탐색기로 열기
explorer .

# 관리자 권한으로 PowerShell 실행
Start-Process powershell -Verb RunAs
```
