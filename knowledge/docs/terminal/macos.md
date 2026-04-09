# macOS Terminal Reference

## Package Management (Homebrew)

### Install/Search/Update
```bash
brew install nodejs
brew install --cask visual-studio-code
brew search python
brew update             # Homebrew 업데이트
brew upgrade            # 패키지 업그레이드
brew upgrade nodejs     # 특정 패키지만
brew list               # 설치된 패키지
brew info nodejs        # 패키지 정보
```

## File Operations

### Open Files/URLs
```bash
open file.txt           # 기본 앱으로 열기
open -a Safari index.html
open .                  # Finder로 현재 디렉토리
open https://example.com
```

### Clipboard
```bash
echo "hello" | pbcopy   # 클립보드에 복사
pbpaste                 # 클립보드 내용 붙여넣기
pbpaste > file.txt
```

## System Settings (defaults)

### Read/Write Settings
```bash
defaults read com.apple.dock
defaults write com.apple.dock autohide -bool true
defaults write com.apple.finder AppleShowAllFiles -bool true
killall Finder          # Finder 재시작
```

### Common Settings
```bash
# Dock 자동 숨김
defaults write com.apple.dock autohide -bool true

# 숨김 파일 표시
defaults write com.apple.finder AppleShowAllFiles -bool true

# 스크린샷 그림자 제거
defaults write com.apple.screencapture disable-shadow -bool true
```

## Network

### Network Setup
```bash
networksetup -getinfo Wi-Fi
networksetup -listallnetworkservices
networksetup -setairportpower en0 on
```

### DNS Cache Flush
```bash
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

## Spotlight Search

### mdfind
```bash
mdfind "query"
mdfind -name file.txt
mdfind "kMDItemKind == 'PDF'"
mdfind -onlyin ~/Documents query
```

## Utilities

### Prevent Sleep (caffeinate)
```bash
caffeinate -t 3600      # 1시간 동안 슬립 방지
caffeinate -d           # 디스플레이 켜진 상태 유지
caffeinate make build   # 명령 실행 중 슬립 방지
```

### Disk Utility (diskutil)
```bash
diskutil list
diskutil info disk0
diskutil eject /Volumes/USB
diskutil verifyDisk disk0
diskutil repairDisk disk0
```

### Software Update
```bash
softwareupdate -l       # 업데이트 목록
softwareupdate -ia      # 모든 업데이트 설치
softwareupdate -i Safari
```

## 자주 쓰는 macOS 명령어

### 파일 경로 복사
```bash
pwd | pbcopy            # 현재 경로 클립보드에 복사
```

### Quick Look 프리뷰
```bash
qlmanage -p file.pdf &> /dev/null
```

### 시스템 정보
```bash
system_profiler SPHardwareDataType
sw_vers                 # macOS 버전
```

### 포트 사용 확인
```bash
lsof -i :8080
sudo lsof -iTCP -sTCP:LISTEN -n -P
```

### 앱 강제 종료
```bash
killall Safari
pkill -f "Google Chrome"
```
