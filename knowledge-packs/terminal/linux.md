# Linux/Bash Terminal Reference

## File Operations

### List Files
```bash
ls
ls -la              # 모든 파일, 상세 정보
ls -lh              # 사람이 읽기 쉬운 크기
ls -R               # 재귀적
```

### Read Files
```bash
cat file.txt
cat file1.txt file2.txt  # 여러 파일
head -n 10 file.txt      # 처음 10줄
tail -n 10 file.txt      # 마지막 10줄
tail -f /var/log/app.log # 실시간 로그
less file.txt            # 페이지 단위
```

### Find & Search
```bash
find . -name "*.log"
find . -type f -size +100M
grep "error" file.txt
grep -r "TODO" .         # 재귀 검색
grep -i "error" file.txt # 대소문자 무시
```

## Network

### IP Address
```bash
ip addr
ip addr show eth0
ifconfig             # 구버전
```

### Network Connections
```bash
ss -tuln             # 리스닝 포트
ss -tunap            # 모든 연결
netstat -tuln        # 구버전
```

### HTTP Request
```bash
curl https://api.example.com/data
curl -X POST -H "Content-Type: application/json" -d '{"key":"value"}' https://api.example.com
wget https://example.com/file.zip
```

## Process Management

### List Processes
```bash
ps aux
ps aux | grep nginx
top                  # 실시간 모니터링
htop                 # 개선된 top
```

### Kill Process
```bash
kill 1234            # PID로 종료
kill -9 1234         # 강제 종료
killall nginx        # 이름으로 종료
pkill -f "python app.py"
```

## Service Management (systemd)

### Start/Stop/Restart
```bash
systemctl start nginx
systemctl stop nginx
systemctl restart nginx
systemctl reload nginx
systemctl status nginx
```

### Enable/Disable (부팅 시 자동 시작)
```bash
systemctl enable nginx
systemctl disable nginx
systemctl is-enabled nginx
```

### Logs
```bash
journalctl -u nginx
journalctl -u nginx -f           # 실시간
journalctl -u nginx --since today
journalctl -xe                   # 최근 에러
```

## Permissions

### Change Permissions (chmod)
```bash
chmod 755 file.sh    # rwxr-xr-x
chmod +x file.sh     # 실행 권한 추가
chmod -R 644 dir/    # 재귀적
```

### Change Owner (chown)
```bash
chown user:group file.txt
chown -R www-data:www-data /var/www
```

## Archive & Compression

### tar
```bash
tar -czf archive.tar.gz dir/     # 압축
tar -xzf archive.tar.gz          # 압축 해제
tar -xzf archive.tar.gz -C /tmp  # 특정 디렉토리에 해제
tar -tzf archive.tar.gz          # 내용 보기
```

## Disk Usage

### df (Disk Free)
```bash
df -h                # 디스크 사용량
df -h /home
```

### du (Disk Usage)
```bash
du -sh dir/          # 디렉토리 크기
du -sh *             # 현재 디렉토리 항목별 크기
du -h --max-depth=1
```

## Package Management

### apt (Debian/Ubuntu)
```bash
apt update
apt upgrade
apt install nginx
apt remove nginx
apt search nodejs
apt list --installed
```

### dnf (Fedora/RHEL)
```bash
dnf update
dnf install nginx
dnf remove nginx
dnf search nodejs
```

## 자주 쓰는 패턴

### 파일 찾기 + 실행
```bash
find . -name "*.log" -exec rm {} \;
find . -type f -mtime +30 -delete
```

### 포트 사용 프로세스 찾기
```bash
ss -tlnp | grep :8080
lsof -i :8080
```

### 디스크 공간 확보
```bash
journalctl --vacuum-time=7d
apt autoremove
apt autoclean
```
