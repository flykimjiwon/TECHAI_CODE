# macOS 개발 환경 Reference

> macOS (Intel / Apple Silicon M1/M2/M3/M4) 개발 환경 설정과 자주 쓰는 명령어 모음.
> 한국어 요청 → macOS 명령어 변환을 위해 택가이코드 LLM이 참조하는 문서.

---

## Homebrew 패키지 관리

macOS의 사실상 표준 패키지 매니저. `/opt/homebrew` (Apple Silicon) 또는 `/usr/local` (Intel)에 설치됨.

```bash
# Homebrew 설치 (미설치 시)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 설치 후 PATH 추가 (Apple Silicon — ~/.zprofile에 추가됨)
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/opt/homebrew/bin/brew shellenv)"

# 패키지 설치
brew install git
brew install node
brew install go
brew install python
brew install wget curl jq

# GUI 앱 설치 (Cask)
brew install --cask visual-studio-code
brew install --cask docker
brew install --cask iterm2

# 패키지 검색
brew search git
brew info git                          # 상세 정보, 의존성 확인

# 설치된 패키지 목록
brew list
brew list --cask                       # GUI 앱 목록

# 업데이트
brew update                            # Homebrew 자체 업데이트
brew upgrade                           # 모든 패키지 업그레이드
brew upgrade git                       # 특정 패키지만

# 제거
brew uninstall git
brew autoremove                        # 불필요한 의존성 제거
brew cleanup                           # 구버전 파일 정리

# 진단
brew doctor                            # 문제 진단
brew --version
```

---

## Xcode Command Line Tools

많은 개발 도구의 전제 조건. macOS 재설치 후 가장 먼저 설치.

```bash
# 설치
xcode-select --install
# 팝업 창이 뜨면 "설치" 클릭

# 설치 확인
xcode-select -p                        # 경로 출력 (보통 /Library/Developer/CommandLineTools)
xcode-select --version

# 재설치 (문제 발생 시)
sudo rm -rf /Library/Developer/CommandLineTools
xcode-select --install

# 전체 Xcode 설치 (iOS 개발 필요 시 — App Store에서)
# xcode-select -s /Applications/Xcode.app  # 전체 Xcode를 CLT로 사용
```

---

## 개발 도구 설치

```bash
# 기본 개발 도구 일괄 설치
brew install git node go python ruby

# Node.js 버전 관리 (fnm 권장 — nvm보다 빠름)
brew install fnm
# ~/.zshrc에 추가
echo 'eval "$(fnm env --use-on-cd)"' >> ~/.zshrc
source ~/.zshrc
fnm install 20                         # Node 20 설치
fnm use 20
fnm default 20                         # 기본 버전 설정

# 또는 nvm 사용
brew install nvm
echo 'export NVM_DIR="$HOME/.nvm"' >> ~/.zshrc
echo '[ -s "/opt/homebrew/opt/nvm/nvm.sh" ] && \. "/opt/homebrew/opt/nvm/nvm.sh"' >> ~/.zshrc
source ~/.zshrc
nvm install 20
nvm use 20

# Go 설치 후 PATH 확인
go env GOPATH                          # 보통 ~/go
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
```

---

## macOS 보안 설정

### Gatekeeper (서명되지 않은 앱 실행)

```bash
# "알 수 없는 개발자" 앱 실행 허용 — 시스템 설정 > 개인 정보 보호 및 보안 에서 허용 클릭
# 또는 터미널로 특정 앱만 허용
spctl --add /Applications/SomeApp.app
xattr -d com.apple.quarantine /Applications/SomeApp.app  # 격리 속성 제거

# Gatekeeper 전체 비활성화 (보안 위험, 비권장)
# sudo spctl --master-disable

# Gatekeeper 상태 확인
spctl --status                         # assessments enabled/disabled

# 바이너리 격리 해제 (직접 다운로드한 실행파일)
xattr -d com.apple.quarantine ./택가이코드
# 또는
xattr -cr ./택가이코드                  # 모든 확장 속성 제거
```

---

## launchctl 서비스 관리 (vs Linux systemd)

| Linux (systemd) | macOS (launchctl) |
|----------------|------------------|
| `systemctl start nginx` | `brew services start nginx` |
| `systemctl stop nginx` | `brew services stop nginx` |
| `systemctl restart nginx` | `brew services restart nginx` |
| `systemctl enable nginx` | `brew services start nginx` (자동 포함) |
| `systemctl status nginx` | `brew services info nginx` |
| `systemctl list-units` | `brew services list` |

```bash
# Homebrew 서비스 관리 (권장)
brew services list                     # 서비스 전체 목록 + 상태
brew services start postgresql@14
brew services stop postgresql@14
brew services restart nginx

# launchctl 직접 사용 (고급)
launchctl list                         # 실행 중인 서비스 목록
launchctl list | grep homebrew        # Homebrew 서비스만
sudo launchctl load /Library/LaunchDaemons/com.example.service.plist
sudo launchctl unload /Library/LaunchDaemons/com.example.service.plist

# LaunchAgent (사용자 서비스) — ~/Library/LaunchAgents/
# LaunchDaemon (시스템 서비스) — /Library/LaunchDaemons/ (관리자)
```

---

## 키체인 관리

```bash
# 키체인에 패스워드 저장 (CI/CD, 스크립트 자동화 시 유용)
security add-generic-password -a "username" -s "service-name" -w "password"

# 키체인에서 패스워드 읽기
security find-generic-password -a "username" -s "service-name" -w

# 인터넷 패스워드 (URL 기반)
security add-internet-password -a "username" -s "example.com" -w "password"
security find-internet-password -a "username" -s "example.com" -w

# 키체인 목록
security list-keychains

# 키체인 잠금/잠금해제
security lock-keychain ~/Library/Keychains/login.keychain-db
security unlock-keychain ~/Library/Keychains/login.keychain-db
```

---

## Spotlight 검색 (mdfind)

```bash
# 파일 이름으로 검색 (find보다 빠름 — 인덱스 기반)
mdfind -name "package.json"
mdfind -name "*.go" -onlyin ~/projects

# 내용으로 검색
mdfind "검색어"
mdfind -onlyin ~/projects "TODO:"

# 메타데이터로 검색
mdfind "kMDItemDisplayName == '*.md'"
mdfind "kMDItemContentType == 'public.go-source'"

# Spotlight 인덱스 재구축 (느려졌을 때)
sudo mdutil -E /               # 전체 재인덱싱
sudo mdutil -a -i on           # 인덱싱 활성화
sudo mdutil -s /               # 인덱싱 상태 확인
```

---

## 디스크 관리 (diskutil)

```bash
# 디스크 목록
diskutil list                          # 전체 디스크 및 파티션
diskutil info /dev/disk0               # 특정 디스크 상세 정보

# 디스크 사용량
df -h                                  # 파일시스템 사용량
du -sh ~/Downloads                     # 특정 폴더 크기
du -sh * | sort -hr | head -20        # 큰 항목 상위 20개

# APFS 볼륨 관리
diskutil apfs list                     # APFS 컨테이너 목록
diskutil apfs addVolume disk1 APFS "NewVolume"  # 볼륨 추가

# 외부 드라이브 마운트/언마운트
diskutil mount /dev/disk2s1
diskutil unmount /dev/disk2s1
diskutil eject /dev/disk2              # 안전 제거

# 디스크 검사/복구
diskutil verifyDisk /dev/disk0
diskutil repairDisk /dev/disk0         # 복구 (외부 디스크)
diskutil verifyVolume /                # 볼륨 검사
```

---

## 파일시스템 차이 (APFS)

```bash
# macOS 기본: APFS (Apple File System)
# 기본적으로 대소문자 비구분 (case-insensitive)
# → Git 저장소에서 대소문자만 다른 파일명 충돌 발생 가능

# 대소문자 구분 여부 확인
diskutil info / | grep "Case Sensitive"

# Git에서 대소문자 변경 추적하게 설정
git config core.ignorecase false

# 대소문자 구분 APFS 볼륨 생성 (개발 전용 파티션)
diskutil apfs addVolume disk1 "APFS (Case-sensitive)" "DevVolume"

# .DS_Store 파일 (Finder 메타데이터) 처리
# 전역 .gitignore에 추가
echo ".DS_Store" >> ~/.gitignore_global
echo "**/.DS_Store" >> ~/.gitignore_global
git config --global core.excludesfile ~/.gitignore_global

# .DS_Store 파일 일괄 삭제
find . -name ".DS_Store" -delete
```

---

## 환경변수 영구 설정

macOS Catalina(10.15) 이후 기본 셸이 bash에서 **zsh**로 변경됨.

```bash
# zsh (macOS 기본) — ~/.zshrc (대화형 셸)
echo 'export MY_VAR="my_value"' >> ~/.zshrc
echo 'export PATH="$PATH:/new/tool/bin"' >> ~/.zshrc
source ~/.zshrc                        # 즉시 적용

# 로그인 셸 전용 — ~/.zprofile
# (터미널 앱 열릴 때, 또는 ssh 로그인 시)
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile

# bash 사용 시 (비권장이지만 호환성 목적)
echo 'export MY_VAR="my_value"' >> ~/.bash_profile  # 로그인 셸
echo 'export MY_VAR="my_value"' >> ~/.bashrc         # 대화형 셸

# 현재 셸 확인
echo $SHELL                            # /bin/zsh 또는 /bin/bash
$SHELL --version

# 셸 변경
chsh -s /bin/zsh                       # zsh로 변경
chsh -s /bin/bash                      # bash로 변경
```

---

## 네트워크

```bash
# 네트워크 인터페이스 목록
networksetup -listallnetworkservices
ifconfig                               # 상세 인터페이스 정보
ip addr                                # Linux 스타일 (iproute2 설치 시)

# Wi-Fi 관련
networksetup -getairportnetwork en0    # 연결된 Wi-Fi 이름
networksetup -setairportpower en0 off # Wi-Fi 끄기
networksetup -setairportpower en0 on  # Wi-Fi 켜기

# DNS 설정
networksetup -getdnsservers Wi-Fi
networksetup -setdnsservers Wi-Fi 8.8.8.8 8.8.4.4  # DNS 변경

# DNS 캐시 초기화 (자주 씀)
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder        # mDNS 재시작

# 포트 확인
lsof -i :8080                          # 8080 포트 사용 프로세스
lsof -i TCP:3000                       # TCP 3000번
netstat -an | grep LISTEN              # 모든 열린 포트

# 라우팅 테이블
netstat -rn
route get 8.8.8.8                      # 특정 IP 라우팅 경로
```

---

## 개발자 성능 팁

```bash
# caffeinate — 슬립 방지 (빌드/다운로드 중 화면 꺼짐 방지)
caffeinate                             # 무한정 슬립 방지 (Ctrl+C로 종료)
caffeinate -t 3600                     # 1시간 동안 슬립 방지
caffeinate -i npm run build            # 명령어 실행 동안만 슬립 방지

# 메모리 압력 해소 (가상 메모리 정리)
sudo purge                             # 파일 캐시 비우기 (⚠️ 일시적으로 느려짐)

# 숨김 파일 표시 (Finder에서)
defaults write com.apple.finder AppleShowAllFiles true
killall Finder                         # Finder 재시작

# 숨김 파일 숨기기 (원래대로)
defaults write com.apple.finder AppleShowAllFiles false
killall Finder

# Dock 자동 숨기기 지연 시간 없애기
defaults write com.apple.dock autohide-delay -float 0
defaults write com.apple.dock autohide-time-modifier -float 0
killall Dock

# 원래대로 복구
defaults delete com.apple.dock autohide-delay
defaults delete com.apple.dock autohide-time-modifier
killall Dock

# 스크린샷 저장 위치 변경
defaults write com.apple.screencapture location ~/Desktop/Screenshots

# 클립보드 내용 확인
pbpaste                                # 클립보드 내용 출력
echo "복사할 텍스트" | pbcopy         # 클립보드에 복사
cat 파일명 | pbcopy                   # 파일 내용을 클립보드에 복사

# 현재 위치를 Finder에서 열기
open .
open -a "Visual Studio Code" .        # VS Code로 열기

# 파일/URL 기본 앱으로 열기
open index.html                        # 브라우저로 열기
open https://example.com              # 브라우저로 URL 열기
```

---

## Apple Silicon (M1/M2/M3/M4) 특이사항

```bash
# 아키텍처 확인
uname -m                               # arm64 (Apple Silicon) 또는 x86_64 (Intel)
arch                                   # arm64 또는 i386

# Rosetta 2 (x86_64 앱을 ARM에서 실행)
# 자동 설치 제안됨. 수동 설치:
softwareupdate --install-rosetta --agree-to-license

# x86_64 모드로 명령 실행 (Rosetta)
arch -x86_64 brew install 패키지       # x86 Homebrew로 설치
arch -x86_64 /usr/local/bin/node      # x86 Node 실행

# Homebrew 위치
# Apple Silicon: /opt/homebrew/
# Intel (또는 Rosetta): /usr/local/
which brew                             # 현재 사용 중인 brew 위치

# Go 크로스 컴파일 (Mac에서 Linux 바이너리 생성)
GOOS=linux GOARCH=amd64 go build -o myapp-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o myapp-linux-arm64 .
GOOS=windows GOARCH=amd64 go build -o myapp-windows-amd64.exe .
```
