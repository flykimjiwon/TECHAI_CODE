# 크로스 플랫폼 명령어 매핑

> **소형 LLM 대상 핵심 참조 문서** — 한국어 자연어 요청을 OS별 실제 명령어로 변환하기 위한 매핑 테이블.
> 에어갭 환경에서 Windows/Linux/macOS 단일 바이너리로 동작하는 택가이코드가 올바른 명령어를 제안하려면 이 문서를 우선 참조해야 함.

---

## 파일/디렉토리

| 한국어 의도 | Windows CMD | Windows PowerShell | Linux | macOS |
|------------|------------|-------------------|-------|-------|
| "D드라이브로 이동해봐" | `D:` | `Set-Location D:\` | _(드라이브 문자 없음 — 마운트 포인트 사용: `cd /mnt/d`)_ | _(동일, N/A)_ |
| "현재 경로 확인" | `cd` | `Get-Location` | `pwd` | `pwd` |
| "상위 폴더로 이동" | `cd ..` | `cd ..` | `cd ..` | `cd ..` |
| "홈 디렉토리로 이동" | `cd %USERPROFILE%` | `cd ~` | `cd ~` | `cd ~` |
| "폴더 만들어" | `mkdir 폴더명` | `New-Item -ItemType Directory 폴더명` | `mkdir -p 폴더명` | `mkdir -p 폴더명` |
| "파일 삭제해" | `del 파일명` | `Remove-Item 파일명` | `rm 파일명` | `rm 파일명` |
| "폴더 삭제해 (비어있지 않아도)" | `rmdir /s /q 폴더명` | `Remove-Item -Recurse -Force 폴더명` | `rm -rf 폴더명` | `rm -rf 폴더명` |
| "파일 복사해" | `copy 원본 대상` | `Copy-Item 원본 대상` | `cp 원본 대상` | `cp 원본 대상` |
| "폴더 통째로 복사해" | `xcopy /s /e 원본 대상` | `Copy-Item -Recurse 원본 대상` | `cp -r 원본 대상` | `cp -r 원본 대상` |
| "파일 이동해" | `move 원본 대상` | `Move-Item 원본 대상` | `mv 원본 대상` | `mv 원본 대상` |
| "파일 내용 봐" | `type 파일명` | `Get-Content 파일명` | `cat 파일명` | `cat 파일명` |
| "파일 끝 10줄만 봐" | _(기본 없음, PowerShell 필요)_ | `Get-Content 파일 -Tail 10` | `tail -n 10 파일` | `tail -n 10 파일` |
| "숨김 파일 보여줘" | `dir /ah` | `Get-ChildItem -Force` | `ls -la` | `ls -la` |
| "파일 이름으로 찾아줘" | `dir /s /b *이름*` | `Get-ChildItem -Recurse -Filter *이름*` | `find . -name "*이름*"` | `find . -name "*이름*"` |
| "내용으로 검색해" | `findstr "텍스트" *.txt` | `Select-String -Pattern "텍스트" -Path *.txt` | `grep -r "텍스트" .` | `grep -r "텍스트" .` |
| "파일 크기 확인" | `dir 파일명` | `(Get-Item 파일명).Length` | `ls -lh 파일명` | `ls -lh 파일명` |
| "폴더 용량 확인" | `du` _(WSL 또는 별도 도구)_ | `(Get-ChildItem -Recurse \| Measure-Object -Property Length -Sum).Sum` | `du -sh 폴더명` | `du -sh 폴더명` |
| "빈 파일 만들어" | `type nul > 파일명` | `New-Item -ItemType File 파일명` | `touch 파일명` | `touch 파일명` |
| "현재 폴더 파일 목록" | `dir` | `Get-ChildItem` (또는 `ls`) | `ls` | `ls` |

---

## 프로세스/네트워크

| 한국어 의도 | Windows CMD | Windows PowerShell | Linux | macOS |
|------------|------------|-------------------|-------|-------|
| "이 포트 뭐가 쓰고있어" | `netstat -ano \| findstr :8080` | `netstat -ano \| findstr :8080` | `ss -tlnp \| grep :8080` | `lsof -i :8080` |
| "포트 점유 프로세스 죽여" | `taskkill /PID 1234 /F` | `Stop-Process -Id 1234 -Force` | `kill -9 1234` | `kill -9 1234` |
| "모든 프로세스 보여줘" | `tasklist` | `Get-Process` | `ps aux` | `ps aux` |
| "특정 프로세스 찾아" | `tasklist \| findstr 이름` | `Get-Process -Name 이름` | `ps aux \| grep 이름` | `ps aux \| grep 이름` |
| "내 IP 확인" | `ipconfig` | `Get-NetIPAddress` | `ip addr` | `ifconfig` |
| "인터넷 연결 확인" | `ping 8.8.8.8` | `Test-Connection 8.8.8.8` | `ping -c 4 8.8.8.8` | `ping -c 4 8.8.8.8` |
| "DNS 캐시 초기화" | `ipconfig /flushdns` | `Clear-DnsClientCache` | `systemd-resolve --flush-caches` | `dscacheutil -flushcache` |
| "서비스 재시작" | `net stop 이름 && net start 이름` | `Restart-Service 이름` | `systemctl restart 이름` | `brew services restart 이름` |
| "서비스 상태 확인" | `sc query 이름` | `Get-Service 이름` | `systemctl status 이름` | `brew services list` |

---

## 환경변수

| 한국어 의도 | Windows CMD | Windows PowerShell | Linux/macOS |
|------------|------------|-------------------|------------|
| "환경변수 설정해 (임시)" | `set VAR=값` | `$env:VAR = "값"` | `export VAR=값` |
| "환경변수 확인해" | `echo %VAR%` | `$env:VAR` | `echo $VAR` |
| "모든 환경변수 보여줘" | `set` | `Get-ChildItem Env:` | `env` 또는 `printenv` |
| "PATH에 추가해 (임시)" | `set PATH=%PATH%;C:\새경로` | `$env:PATH += ";C:\새경로"` | `export PATH="$PATH:/새경로"` |
| "환경변수 영구 설정" | `setx VAR "값"` | `[System.Environment]::SetEnvironmentVariable("VAR","값","User")` | `echo 'export VAR=값' >> ~/.bashrc && source ~/.bashrc` |

---

## 개발 도구 설치

| 한국어 의도 | Windows | Ubuntu/Debian | Fedora/RHEL | macOS |
|------------|---------|--------------|------------|-------|
| "Git 설치해" | `winget install Git.Git` | `sudo apt install git` | `sudo dnf install git` | `brew install git` |
| "Node.js 설치해" | `winget install OpenJS.NodeJS.LTS` | `sudo apt install nodejs npm` | `sudo dnf install nodejs` | `brew install node` |
| "Go 설치해" | `winget install GoLang.Go` | `sudo apt install golang-go` | `sudo dnf install golang` | `brew install go` |
| "Python 설치해" | `winget install Python.Python.3` | `sudo apt install python3 python3-pip` | `sudo dnf install python3` | `brew install python` |
| "Docker 설치해" | winget으로 Docker Desktop 설치 | `sudo apt install docker.io` | `sudo dnf install docker` | `brew install --cask docker` |
| "패키지 검색해" | `winget search 이름` | `apt search 이름` | `dnf search 이름` | `brew search 이름` |

---

## Node.js 버전 관리

| 한국어 의도 | Windows (nvm-windows) | Linux/macOS (nvm) | Linux/macOS (fnm) |
|------------|----------------------|------------------|------------------|
| "Node 버전 변경해" | `nvm use 20` | `nvm use 20` | `fnm use 20` |
| "설치된 버전 보여줘" | `nvm list` | `nvm ls` | `fnm ls` |
| "새 버전 설치해" | `nvm install 20` | `nvm install 20` | `fnm install 20` |
| "기본 버전 설정해" | `nvm alias default 20` (미지원, 별도 방법) | `nvm alias default 20` | `fnm default 20` |

---

## 방화벽

| 한국어 의도 | Windows PowerShell | Ubuntu (ufw) | Fedora (firewalld) | macOS |
|------------|-------------------|-------------|-------------------|-------|
| "포트 열어줘" | `netsh advfirewall firewall add rule name="포트명" dir=in action=allow protocol=TCP localport=8080` | `sudo ufw allow 8080` | `sudo firewall-cmd --add-port=8080/tcp --permanent && sudo firewall-cmd --reload` | _(pf 사용, 일반적으로 앱 방화벽 GUI 사용)_ |
| "포트 막아줘" | `netsh advfirewall firewall add rule name="포트명" dir=in action=block protocol=TCP localport=8080` | `sudo ufw deny 8080` | `sudo firewall-cmd --remove-port=8080/tcp --permanent` | _(동일)_ |
| "방화벽 상태 확인" | `netsh advfirewall show allprofiles` | `sudo ufw status verbose` | `sudo firewall-cmd --list-all` | `sudo pfctl -s rules` |

---

## 스크립트 실행

| 항목 | Windows CMD | Windows PowerShell | Linux | macOS |
|-----|------------|-------------------|-------|-------|
| 스크립트 확장자 | `.bat`, `.cmd` | `.ps1` | `.sh` | `.sh`, `.command` |
| 실행 방법 | `script.bat` | `.\script.ps1` | `./script.sh` _(chmod +x 필요)_ | `./script.sh` |
| 실행 권한 부여 | _(없음, 기본 실행 가능)_ | _(없음, 단 ExecutionPolicy 설정 필요)_ | `chmod +x script.sh` | `chmod +x script.sh` |
| PowerShell 실행 정책 허용 | — | `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned` | — | — |

---

## 경로 차이 요약

| 항목 | Windows | Linux | macOS |
|-----|---------|-------|-------|
| 경로 구분자 | `\` (백슬래시) | `/` (슬래시) | `/` (슬래시) |
| 홈 디렉토리 | `C:\Users\username` | `/home/username` | `/Users/username` |
| 임시 디렉토리 | `%TEMP%` (보통 `C:\Users\username\AppData\Local\Temp`) | `/tmp` | `/tmp` |
| 전역 npm 모듈 | `%APPDATA%\npm\node_modules` | `/usr/local/lib/node_modules` | `/usr/local/lib/node_modules` |
| Go 바이너리 | `%GOPATH%\bin` (보통 `%USERPROFILE%\go\bin`) | `$GOPATH/bin` 또는 `~/go/bin` | `$GOPATH/bin` 또는 `~/go/bin` |
| 줄바꿈 문자 | `\r\n` (CRLF) | `\n` (LF) | `\n` (LF) |
| 대소문자 구분 | 구분 안 함 (기본) | 구분함 | 구분 안 함 (APFS 기본) |
| 실행파일 확장자 | `.exe`, `.bat`, `.ps1` | 확장자 불필요 (실행 비트) | 확장자 불필요 (실행 비트) |

---

## OS 감지 명령어

```bash
# Linux/macOS에서 OS 구분
uname -s          # Darwin(macOS) 또는 Linux 출력
uname -m          # 아키텍처 (x86_64, arm64 등)
cat /etc/os-release  # Linux 배포판 정보 (Ubuntu, Fedora 등)
sw_vers           # macOS 버전 상세 (macOS 전용)
```

```powershell
# Windows PowerShell에서 OS 정보
[System.Environment]::OSVersion  # OS 버전 객체
$PSVersionTable                   # PowerShell 버전 및 플랫폼
(Get-ComputerInfo).OsName         # 상세 OS 이름
```

---

## 소형 LLM 판단 기준

OS를 명시하지 않은 경우 아래 신호로 판단:

| 신호 | 추정 OS |
|-----|--------|
| "D드라이브", "C드라이브", "드라이브" 언급 | Windows |
| `\` 경로 포함 | Windows |
| "brew", "homebrew" 언급 | macOS |
| "apt", "yum", "dnf", "systemctl" 언급 | Linux |
| "Mac", "맥북", "M1", "M2", "M3", "M4" 언급 | macOS |
| "우분투", "데비안", "페도라", "CentOS", "RHEL" 언급 | Linux |
| PowerShell 관련 언급 (`$env:`, `Get-*`) | Windows |
| `/Users/` 경로 | macOS |
| `/home/` 경로 | Linux |
