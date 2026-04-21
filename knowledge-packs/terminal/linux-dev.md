# Linux 개발 환경 Reference

> Ubuntu/Debian, Fedora/RHEL/CentOS, Arch 등 주요 배포판에서 개발할 때 쓰는 명령어 모음.
> 한국어 요청 → Linux 명령어 변환을 위해 택가이코드 LLM이 참조하는 문서.

---

## 패키지 관리

### Ubuntu/Debian (apt)

```bash
# 패키지 목록 업데이트 (설치 전 항상 먼저 실행)
sudo apt update

# 패키지 설치
sudo apt install git nodejs npm golang python3 python3-pip docker.io

# 설치된 패키지 업그레이드
sudo apt upgrade
sudo apt full-upgrade                  # 의존성까지 변경하는 완전 업그레이드

# 패키지 제거 (설정 파일은 유지)
sudo apt remove 패키지명

# 패키지 완전 제거 (설정 파일 포함)
sudo apt purge 패키지명

# 사용하지 않는 의존성 제거
sudo apt autoremove

# 패키지 검색
apt search 검색어
apt show 패키지명                      # 상세 정보
```

### Fedora/RHEL/CentOS (dnf/yum)

```bash
# Fedora 38+ / RHEL 9+
sudo dnf install git nodejs golang python3

# 업데이트
sudo dnf update

# 패키지 검색
dnf search git
dnf info git

# 제거
sudo dnf remove 패키지명

# CentOS 7 / RHEL 7 (구형)
sudo yum install git
sudo yum update
```

### Arch Linux (pacman)

```bash
sudo pacman -Syu                       # 전체 업데이트
sudo pacman -S git nodejs go python    # 설치
sudo pacman -R 패키지명               # 제거
pacman -Ss 검색어                     # 검색
```

---

## 사용자/권한 관리

```bash
# 현재 사용자 확인
whoami
id                                     # UID, GID, 그룹 전체 표시

# sudo — 루트 권한으로 명령 실행
sudo 명령어
sudo -i                                # 루트 셸로 전환
sudo -u 다른사용자 명령어              # 다른 사용자로 실행

# 파일/폴더 권한 변경 (chmod)
chmod 755 파일명                       # rwxr-xr-x (소유자 실행+읽기+쓰기, 나머지 읽기+실행)
chmod 644 파일명                       # rw-r--r-- (일반 파일 기본)
chmod +x 스크립트.sh                   # 실행 권한만 추가
chmod -R 755 폴더명                    # 폴더 내 전체 재귀 변경

# 권한 숫자 계산: r=4, w=2, x=1 → 합산
# 7 = rwx, 6 = rw-, 5 = r-x, 4 = r--

# 파일 소유자 변경 (chown)
sudo chown 사용자명 파일명
sudo chown 사용자명:그룹명 파일명
sudo chown -R 사용자명 폴더명         # 재귀

# 그룹에 사용자 추가 (docker 등 사용 시 필수)
sudo usermod -aG docker $USER
sudo usermod -aG sudo $USER           # sudo 권한 부여
newgrp docker                          # 그룹 변경 즉시 적용 (재로그인 없이)

# 파일 권한/소유자 확인
ls -la 파일명
stat 파일명
```

---

## systemd 서비스 관리

```bash
# 서비스 상태 확인
systemctl status nginx
systemctl status docker

# 서비스 시작/정지/재시작
sudo systemctl start nginx
sudo systemctl stop nginx
sudo systemctl restart nginx
sudo systemctl reload nginx            # 설정만 재로드 (무중단)

# 부팅 시 자동 시작 설정
sudo systemctl enable nginx
sudo systemctl enable --now nginx      # 활성화 + 즉시 시작

# 자동 시작 해제
sudo systemctl disable nginx

# 전체 서비스 목록
systemctl list-units --type=service
systemctl list-units --type=service --state=running  # 실행 중인 것만

# 사용자 서비스 (sudo 없이)
systemctl --user status my-service
systemctl --user enable my-service
```

---

## 파일시스템 구조

| 경로 | 용도 |
|-----|------|
| `/etc` | 시스템 설정 파일 (nginx.conf, ssh/sshd_config 등) |
| `/var/log` | 시스템/앱 로그 파일 |
| `/var/lib` | 앱 데이터 (docker, mysql 등) |
| `/opt` | 수동 설치 타사 소프트웨어 |
| `/usr/local/bin` | 수동 설치 실행파일 |
| `/usr/local/lib` | 수동 설치 라이브러리 |
| `/tmp` | 임시 파일 (재부팅 시 삭제) |
| `/home/username` | 사용자 홈 디렉토리 |
| `/root` | 루트 사용자 홈 |
| `/proc` | 프로세스/커널 가상 파일시스템 |
| `/sys` | 하드웨어/커널 가상 파일시스템 |
| `/dev` | 장치 파일 |
| `/mnt`, `/media` | 마운트 포인트 |

---

## SSH 키 생성 및 관리

```bash
# SSH 키 생성 (Ed25519 권장, RSA 4096도 OK)
ssh-keygen -t ed25519 -C "your@email.com"
ssh-keygen -t rsa -b 4096 -C "your@email.com"
# 기본 저장 위치: ~/.ssh/id_ed25519, ~/.ssh/id_ed25519.pub

# 공개키를 원격 서버에 복사 (패스워드 인증이 가능할 때)
ssh-copy-id username@서버IP
ssh-copy-id -i ~/.ssh/id_ed25519.pub username@서버IP

# 수동으로 공개키 등록 (서버에서)
cat ~/.ssh/id_ed25519.pub >> ~/.ssh/authorized_keys
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys

# SSH 접속 테스트
ssh username@서버IP
ssh -i ~/.ssh/특정키 username@서버IP   # 특정 키 지정

# SSH 설정 파일 (~/.ssh/config) — 별칭 설정
# Host myserver
#     HostName 192.168.1.100
#     User ubuntu
#     IdentityFile ~/.ssh/id_ed25519
# 이후 ssh myserver 만으로 접속 가능

# SSH 에이전트 (패스프레이즈 캐시)
eval $(ssh-agent -s)
ssh-add ~/.ssh/id_ed25519
ssh-add -l                             # 등록된 키 목록
```

---

## 방화벽

### UFW (Ubuntu)

```bash
# UFW 상태 확인
sudo ufw status verbose
sudo ufw status numbered               # 번호 포함

# 활성화/비활성화
sudo ufw enable
sudo ufw disable

# 포트 허용
sudo ufw allow 8080                    # TCP/UDP 모두
sudo ufw allow 8080/tcp
sudo ufw allow ssh                     # 22번 포트 허용 (이름으로 지정)
sudo ufw allow from 192.168.1.0/24    # 특정 IP 대역 허용

# 포트 차단
sudo ufw deny 8080
sudo ufw delete allow 8080            # 규칙 삭제
sudo ufw delete 3                     # 번호로 규칙 삭제
```

### firewalld (Fedora/RHEL)

```bash
# 상태 확인
sudo firewall-cmd --state
sudo firewall-cmd --list-all

# 포트 열기 (즉시 + 영구)
sudo firewall-cmd --add-port=8080/tcp
sudo firewall-cmd --add-port=8080/tcp --permanent  # 재부팅 후에도 유지
sudo firewall-cmd --reload                          # 영구 설정 적용

# 포트 닫기
sudo firewall-cmd --remove-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

---

## 로그 확인

```bash
# journalctl — systemd 서비스 로그
journalctl -u nginx                    # nginx 서비스 로그
journalctl -u nginx -f                 # 실시간 팔로우
journalctl -u nginx --since "1 hour ago"
journalctl -u nginx --since "2024-01-01" --until "2024-01-02"
journalctl -n 100                      # 최근 100줄
journalctl -p err                      # 에러 레벨 이상만 (emerg, alert, crit, err)

# 전통적인 로그 파일
tail -f /var/log/nginx/access.log     # 실시간 접근 로그
tail -f /var/log/nginx/error.log      # 에러 로그
tail -n 100 /var/log/syslog           # 시스템 로그 최근 100줄
grep "ERROR" /var/log/app.log         # 에러 필터
grep "ERROR" /var/log/app.log | tail -50  # 최근 에러 50개
```

---

## cron 작업 (스케줄 실행)

```bash
# crontab 편집
crontab -e                             # 현재 사용자 cron
sudo crontab -e                        # 루트 cron

# cron 표현식: 분 시 일 월 요일 명령어
# 분(0-59) 시(0-23) 일(1-31) 월(1-12) 요일(0-7, 0=일=7)
# *는 "모든 값"

# 예시
0 2 * * *  /home/user/backup.sh           # 매일 새벽 2시
*/5 * * * * /home/user/check.sh           # 5분마다
0 9 * * 1  /home/user/weekly.sh           # 매주 월요일 9시
0 0 1 * *  /home/user/monthly.sh          # 매월 1일 자정
@reboot    /home/user/startup.sh          # 부팅 시 1회

# cron 목록 확인
crontab -l

# cron 로그 확인
journalctl -u cron
grep cron /var/log/syslog
```

---

## 디스크 관리

```bash
# 디스크 사용량 확인
df -h                                  # 전체 파일시스템 사용량 (사람이 읽기 쉬운 형식)
df -h /home                           # 특정 경로의 파일시스템

# 폴더/파일 크기 확인
du -sh 폴더명                          # 폴더 전체 크기
du -sh *                               # 현재 폴더 내 항목별 크기
du -sh * | sort -hr                    # 크기 내림차순 정렬

# 마운트 관리
mount                                  # 마운트된 파일시스템 목록
sudo mount /dev/sdb1 /mnt/data        # 장치 마운트
sudo umount /mnt/data                  # 언마운트

# /etc/fstab — 부팅 시 자동 마운트 설정
# UUID=xxxx  /mnt/data  ext4  defaults  0  2

# 파티션/디스크 정보
lsblk                                  # 블록 장치 목록 (트리 형식)
sudo fdisk -l                          # 상세 파티션 정보 (관리자 필요)
```

---

## 프로세스 관리

```bash
# 프로세스 목록
ps aux                                 # 전체 프로세스 (상세)
ps aux | grep 이름                     # 특정 프로세스 검색
pgrep -a 이름                          # 프로세스 이름으로 PID 찾기

# 실시간 모니터링
top                                    # 기본 (q로 종료)
htop                                   # 개선된 버전 (apt install htop)

# 프로세스 종료
kill PID                               # SIGTERM (정상 종료 요청)
kill -9 PID                            # SIGKILL (강제 종료)
killall 이름                           # 이름으로 전체 종료
pkill 이름                             # 이름 패턴으로 종료

# 백그라운드 실행
명령어 &                               # 백그라운드로 실행
nohup 명령어 &                         # 터미널 종료 후에도 계속 실행
nohup 명령어 > output.log 2>&1 &       # 로그 파일로 출력 저장

# 백그라운드 작업 관리
jobs                                   # 백그라운드 작업 목록
fg %1                                  # 1번 작업을 포그라운드로
bg %1                                  # 1번 작업을 백그라운드로
Ctrl+Z                                 # 현재 작업 일시 중단 (suspend)
```

---

## 환경변수 영구 설정

```bash
# 현재 사용자만 — ~/.bashrc (bash) 또는 ~/.zshrc (zsh)
echo 'export MY_VAR="my_value"' >> ~/.bashrc
echo 'export PATH="$PATH:/new/tool/bin"' >> ~/.bashrc
source ~/.bashrc                       # 즉시 적용

# 로그인 셸용 — ~/.profile 또는 ~/.bash_profile
echo 'export MY_VAR="my_value"' >> ~/.profile

# 시스템 전체 — /etc/environment (재부팅 필요)
sudo bash -c 'echo "MY_VAR=my_value" >> /etc/environment'

# 현재 셸 세션에만 임시 설정
export MY_VAR="my_value"

# 환경변수 확인
echo $MY_VAR
env | grep MY_VAR
printenv MY_VAR
```

---

## Docker 기본

```bash
# 이미지 관리
docker pull ubuntu:22.04               # 이미지 다운로드
docker images                          # 로컬 이미지 목록
docker rmi ubuntu:22.04               # 이미지 삭제

# 컨테이너 실행
docker run -it ubuntu:22.04 bash       # 대화형 실행
docker run -d -p 8080:80 nginx         # 백그라운드 + 포트 매핑
docker run -d -v /host/path:/container/path nginx  # 볼륨 마운트
docker run --rm ubuntu:22.04 echo "hello"  # 실행 후 자동 삭제

# 컨테이너 관리
docker ps                              # 실행 중인 컨테이너
docker ps -a                           # 전체 컨테이너 (중지 포함)
docker stop 컨테이너ID                 # 정상 종료
docker rm 컨테이너ID                   # 컨테이너 삭제
docker exec -it 컨테이너ID bash        # 실행 중인 컨테이너에 접속

# docker-compose
docker compose up -d                   # 백그라운드 실행
docker compose down                    # 중지 + 컨테이너 삭제
docker compose down -v                 # 볼륨까지 삭제
docker compose logs -f                 # 실시간 로그
docker compose ps                      # 컨테이너 상태

# Docker 설치 후 sudo 없이 사용하기
sudo usermod -aG docker $USER
newgrp docker                          # 재로그인 없이 즉시 적용
```
