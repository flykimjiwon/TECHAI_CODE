# Shell Ops Reference

## cron

```bash
crontab -e          # 편집
crontab -l          # 현재 목록
crontab -r          # 삭제
crontab -u user -e  # 특정 유저

# 문법: 분(0-59) 시(0-23) 일(1-31) 월(1-12) 요일(0-7, 0=7=일)
# *      — 매번
# */5    — 5마다
# 1,15   — 1일, 15일
# 1-5    — 1~5

0 * * * *       /script.sh        # 매시 정각
*/15 * * * *    /check.sh         # 15분마다
0 2 * * *       /backup.sh        # 매일 02:00
0 2 * * 0       /weekly.sh        # 매주 일요일 02:00
0 2 1 * *       /monthly.sh       # 매월 1일 02:00
@reboot         /startup.sh       # 부팅 시
@daily          /daily.sh         # 매일 자정 (= 0 0 * * *)
@hourly         /hourly.sh        # 매시 (= 0 * * * *)

# 로그 리다이렉트
0 2 * * * /backup.sh >> /var/log/backup.log 2>&1

# 환경변수 설정 (crontab 상단)
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin
MAILTO=admin@example.com
```

---

## systemd

```bash
# 유닛 파일 위치
/etc/systemd/system/myapp.service      # 시스템 서비스
~/.config/systemd/user/myapp.service   # 유저 서비스

# 기본 서비스 유닛
cat > /etc/systemd/system/myapp.service << 'EOF'
[Unit]
Description=My Application
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=appuser
WorkingDirectory=/opt/myapp
ExecStart=/opt/myapp/bin/server
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal
Environment=PORT=8080
EnvironmentFile=/etc/myapp/env

[Install]
WantedBy=multi-user.target
EOF

# 타이머 유닛 (cron 대체)
cat > /etc/systemd/system/backup.timer << 'EOF'
[Unit]
Description=Daily Backup Timer

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
EOF

# systemctl 명령
systemctl daemon-reload          # 유닛 파일 변경 후 반드시 실행
systemctl enable myapp           # 부팅 시 자동 시작 등록
systemctl disable myapp          # 자동 시작 해제
systemctl start myapp
systemctl stop myapp
systemctl restart myapp
systemctl reload myapp           # 설정 재로드 (재시작 없이)
systemctl status myapp           # 상태 + 최근 로그
systemctl is-active myapp        # active/inactive
systemctl is-enabled myapp       # enabled/disabled
systemctl list-units --type=service --state=running
systemctl list-timers
```

---

## 로그

```bash
# journalctl (systemd 로그)
journalctl -u myapp              # 특정 서비스
journalctl -u myapp -f           # 실시간 (tail -f)
journalctl -u myapp -n 100       # 최근 100줄
journalctl -u myapp --since "1 hour ago"
journalctl -u myapp --since "2024-01-01" --until "2024-01-02"
journalctl -p err                # 에러 이상만 (emerg/alert/crit/err/warning/notice/info/debug)
journalctl --disk-usage
journalctl --vacuum-time=7d      # 7일 이상 삭제

# 파일 로그
tail -f /var/log/syslog
tail -f /var/log/nginx/access.log
grep "ERROR" /var/log/app.log | tail -50

# logrotate 설정 예시
cat > /etc/logrotate.d/myapp << 'EOF'
/var/log/myapp/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    sharedscripts
    postrotate
        systemctl reload myapp
    endscript
}
EOF
```

---

## SSH

```bash
# 키 생성
ssh-keygen -t ed25519 -C "comment" -f ~/.ssh/id_ed25519
ssh-keygen -t rsa -b 4096 -C "comment"

# 공개키 배포
ssh-copy-id -i ~/.ssh/id_ed25519.pub user@host
# 수동
cat ~/.ssh/id_ed25519.pub >> ~/.ssh/authorized_keys

# ~/.ssh/config
cat >> ~/.ssh/config << 'EOF'
Host prod
    HostName 203.0.113.10
    User deploy
    IdentityFile ~/.ssh/id_ed25519
    Port 22

Host bastion
    HostName 203.0.113.1
    User ec2-user
    IdentityFile ~/.ssh/aws.pem
    ForwardAgent yes

Host internal-*
    ProxyJump bastion
    User deploy
EOF
chmod 600 ~/.ssh/config

# 접속
ssh prod
ssh -p 2222 user@host
ssh -i ~/.ssh/key.pem user@host

# 포트 포워딩
ssh -L 8080:localhost:80 user@host      # 로컬 8080 → 원격 80 (로컬 포워딩)
ssh -R 9090:localhost:3000 user@host    # 원격 9090 → 로컬 3000 (리모트 포워딩)
ssh -L 5432:db-internal:5432 bastion    # bastion 경유로 DB 접근

# 원격 명령
ssh user@host "uptime && df -h"
ssh user@host 'bash -s' < local-script.sh
```

---

## 프로세스 관리

```bash
# 조회
ps aux                          # 전체 프로세스
ps aux | grep myapp
pgrep -f myapp                  # PID만
pgrep -a myapp                  # PID + 명령

# 종료
kill PID                        # SIGTERM (정상 종료 요청)
kill -9 PID                     # SIGKILL (강제)
kill -HUP PID                   # SIGHUP (설정 재로드)
pkill -f myapp                  # 이름으로 종료

# 백그라운드 실행
./server &                      # 백그라운드
nohup ./server &                # 터미널 닫아도 유지
nohup ./server > /var/log/server.log 2>&1 &

jobs                            # 현재 셸 백그라운드 목록
fg %1                           # 포그라운드로
bg %1                           # 백그라운드로
disown %1                       # 셸에서 분리 (nohup 효과)

# 우선순위
nice -n 10 ./heavy-task         # 낮은 우선순위로 시작
renice +10 -p PID               # 실행 중 우선순위 변경

# 리소스 모니터
top -b -n1 | head -30           # 일회성 top 출력
htop                            # 대화형 (설치 필요)
```

---

## 디스크

```bash
df -h                           # 파티션별 사용량
df -h /var                      # 특정 경로
df -i                           # inode 사용량

du -sh /var/log                 # 디렉토리 합계
du -sh /var/log/*               # 하위 항목별
du -h --max-depth=2 /var        # 깊이 제한
du -sh * | sort -rh | head -20  # 큰 것 순 정렬

# 실시간 I/O
iostat -x 1                     # 초당 I/O 통계
iotop                           # 프로세스별 I/O (설치 필요)
lsof +D /path                   # 경로 사용 중인 프로세스

# 대용량 파일 찾기
find / -type f -size +1G -ls 2>/dev/null
```

---

## 네트워크

```bash
# 포트 & 소켓
ss -tlnp                        # TCP LISTEN 포트 + 프로세스
ss -tlnp sport = :8080          # 특정 포트
netstat -tlnp                   # ss 없을 때 (deprecated)
lsof -i :8080                   # 포트 사용 프로세스

# 연결 확인
ss -s                           # 연결 요약
ss -tnp state established       # 활성 연결

# DNS
dig example.com                 # A 레코드
dig example.com MX              # MX 레코드
dig @8.8.8.8 example.com        # 특정 DNS 서버
nslookup example.com
host example.com

# 연결 테스트
curl -s -o /dev/null -w "%{http_code}\n" https://example.com
wget -q --spider https://example.com && echo "OK"
nc -zv host 5432                # TCP 포트 연결 확인
telnet host 25                  # 포트 테스트
```

---

## 사용자 & 권한

```bash
# chmod
chmod 755 file          # rwxr-xr-x
chmod 644 file          # rw-r--r--
chmod +x script.sh      # 실행 권한 추가
chmod -R 755 dir/       # 재귀
chmod u+x,g-w file      # 심볼릭

# chown
chown user file
chown user:group file
chown -R user:group dir/

# umask (기본 권한 마스크)
umask              # 현재 값 (보통 022)
umask 027          # 그룹 쓰기, 기타 전부 차단
# 파일 기본: 666 - umask = 644 (umask 022)
# 디렉토리: 777 - umask = 755

# sudo
sudo command
sudo -u appuser command         # 다른 유저로
sudo -E command                 # 환경변수 유지
visudo                          # sudoers 편집 (검증 포함)

# /etc/sudoers.d/myapp
# deploy ALL=(ALL) NOPASSWD: /bin/systemctl restart myapp
```

---

## 환경변수

```bash
# 우선순위: 셸 변수 > .bashrc/.zshrc > .bash_profile/.profile > /etc/environment

# /etc/environment     — 시스템 전체 (key=value 형식, export 없음)
# /etc/profile.d/*.sh  — 로그인 셸 전체 (스크립트)
# ~/.bash_profile       — 로그인 셸, 사용자
# ~/.bashrc             — 인터랙티브 셸, 사용자
# ~/.profile            — POSIX (bash_profile 없을 때 사용)

# .bashrc vs .bash_profile
# .bash_profile: ssh 로그인, 터미널 앱 첫 실행 → PATH 설정
# .bashrc: 새 터미널 탭, bash 호출 → alias, 함수, PS1

# .bash_profile에서 .bashrc 로드 (일반적 패턴)
[[ -f ~/.bashrc ]] && source ~/.bashrc

# 즉시 적용
source ~/.bashrc
. ~/.bashrc        # POSIX 방식

# 실전
export PATH="$PATH:/opt/myapp/bin"
export JAVA_HOME=/usr/lib/jvm/java-17
```

---

## 배포 스크립트 패턴

```bash
# rsync 배포
rsync -avz --delete \
  --exclude='.git' \
  --exclude='node_modules' \
  ./dist/ user@prod:/opt/myapp/

# rsync + SSH 옵션
rsync -avz -e "ssh -i ~/.ssh/deploy.pem -p 22" \
  ./dist/ deploy@prod:/opt/myapp/

# blue-green 기본 패턴
deploy_blue_green() {
  local ACTIVE=$(readlink /opt/app/current)   # current → blue 또는 green
  local NEXT=$([[ "$ACTIVE" == *blue* ]] && echo "green" || echo "blue")

  rsync -az ./dist/ "/opt/app/$NEXT/"
  systemctl start "myapp-$NEXT"

  # 헬스 체크
  for i in {1..10}; do
    curl -sf "http://localhost:8081/health" && break
    sleep 2
  done

  ln -sfn "/opt/app/$NEXT" /opt/app/current
  systemctl stop "myapp-${ACTIVE##*/}"
}
```

---

## 실전 스크립트 모음

```bash
# 서버 초기화 (Ubuntu)
#!/usr/bin/env bash
set -euo pipefail

init_server() {
  apt-get update -qq
  apt-get install -y --no-install-recommends \
    curl git vim htop jq unzip ca-certificates

  # 타임존
  timedatectl set-timezone Asia/Seoul

  # 스왑 (RAM 2GB 서버용)
  [[ ! -f /swapfile ]] && {
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
  }
}

# 헬스체크 스크립트
#!/usr/bin/env bash
check_health() {
  local url="${1:?URL required}"
  local max_retry="${2:-5}"
  local wait="${3:-3}"

  for i in $(seq 1 "$max_retry"); do
    HTTP=$(curl -sf -o /dev/null -w "%{http_code}" "$url") && {
      echo "OK ($HTTP)"
      return 0
    }
    echo "시도 $i/$max_retry 실패 (${HTTP:-timeout}), ${wait}초 후 재시도..."
    sleep "$wait"
  done

  echo "헬스체크 실패: $url" >&2
  return 1
}

check_health "http://localhost:8080/health"

# 로그 로테이션 (logrotate 없을 때 수동)
rotate_log() {
  local log_file="$1"
  local max_files="${2:-7}"

  [[ -f "$log_file" ]] || return 0

  for i in $(seq $((max_files-1)) -1 1); do
    [[ -f "${log_file}.$i" ]] && mv "${log_file}.$i" "${log_file}.$((i+1))"
  done

  mv "$log_file" "${log_file}.1"
  touch "$log_file"
  # 앱에 SIGHUP 전달
  pkill -HUP -f myapp 2>/dev/null || true
}
```
