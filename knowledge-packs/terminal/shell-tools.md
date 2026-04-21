# Shell Tools Reference

## grep

```bash
grep "pattern" file.txt
grep -r "pattern" ./dir        # 재귀 탐색
grep -i "pattern" file         # 대소문자 무시
grep -l "pattern" *.txt        # 매칭 파일명만 출력
grep -L "pattern" *.txt        # 매칭 안된 파일명
grep -n "pattern" file         # 줄 번호 포함
grep -v "pattern" file         # 역매칭 (패턴 없는 줄)
grep -c "pattern" file         # 매칭 줄 수 출력
grep -o "pattern" file         # 매칭 부분만 출력
grep -A 3 "pattern" file       # 매칭 후 3줄
grep -B 3 "pattern" file       # 매칭 전 3줄
grep -C 3 "pattern" file       # 앞뒤 3줄

# 정규식
grep -E "err(or)?" file        # ERE (extended regex)
grep -P "\d{3}-\d{4}" file     # PCRE (Perl-compatible, GNU grep)
grep "^ERROR" file             # 줄 시작
grep "done$" file              # 줄 끝
grep "foo.*bar" file           # foo 이후 bar

# 실전
grep -rn "TODO" --include="*.go" .
grep -v "^#\|^$" /etc/ssh/sshd_config   # 주석/빈줄 제거
ps aux | grep -v grep | grep nginx
```

---

## sed

```bash
# 기본 치환: s/패턴/교체/플래그
sed 's/foo/bar/' file           # 첫 번째만
sed 's/foo/bar/g' file          # 전부 (global)
sed 's/foo/bar/gi' file         # 대소문자 무시 + 전부
sed 's/foo/bar/2' file          # 두 번째만

# 인플레이스 수정
sed -i 's/foo/bar/g' file       # Linux
sed -i '' 's/foo/bar/g' file    # macOS (백업 없음)
sed -i.bak 's/foo/bar/g' file   # 백업 .bak 생성

# 줄 삭제
sed '/pattern/d' file           # 패턴 포함 줄 삭제
sed '3d' file                   # 3번째 줄 삭제
sed '3,7d' file                 # 3~7번째 줄 삭제
sed '/^$/d' file                # 빈 줄 삭제
sed '/^#/d' file                # 주석 줄 삭제

# 주소 범위
sed '5,10s/foo/bar/g' file      # 5~10번째 줄만 치환
sed '/START/,/END/d' file       # START~END 구간 삭제

# 줄 출력/추가
sed -n '5,10p' file             # 5~10번째 줄만 출력 (-n: 기본 출력 억제)
sed '3a\새 줄 내용' file        # 3번째 줄 뒤에 추가
sed '3i\새 줄 내용' file        # 3번째 줄 앞에 삽입

# 여러 표현식
sed -e 's/foo/bar/g' -e 's/baz/qux/g' file

# 실전
sed 's/#.*//' file              # 인라인 주석 제거
sed 's/[[:space:]]*$//' file    # 줄 끝 공백 제거
sed 's/^[[:space:]]*//' file    # 줄 앞 공백 제거
```

---

## awk

```bash
# 기본: awk 'pattern { action }' file
awk '{ print $1 }' file         # 첫 번째 필드
awk '{ print $1, $3 }' file     # 1, 3번째 필드
awk '{ print NR, $0 }' file     # 줄 번호 + 전체 줄
awk 'NR==3' file                # 3번째 줄
awk 'NR>=5 && NR<=10' file      # 5~10번째 줄
awk '/pattern/' file            # 패턴 매칭 줄

# 필드 구분자
awk -F: '{ print $1 }' /etc/passwd      # : 구분
awk -F',' '{ print $2 }' data.csv       # , 구분
awk 'BEGIN { FS="," } { print $2 }' file

# 변수
awk -v threshold=100 '$3 > threshold { print }' file

# BEGIN / END
awk 'BEGIN { print "시작" } { print } END { print "종료" }' file

# 조건
awk '$3 > 100 { print $1, $3 }' file
awk '$1 == "ERROR" { print $0 }' file
awk 'NF > 3 { print }' file             # 필드 3개 초과인 줄

# 합계 / 평균
awk '{ sum += $3 } END { print "합계:", sum }' file
awk '{ sum += $3; cnt++ } END { print "평균:", sum/cnt }' file

# 출력 포맷
awk '{ printf "%-20s %5d\n", $1, $2 }' file

# 연관 배열 (집계)
awk '{ count[$1]++ } END { for (k in count) print k, count[k] }' log.txt

# 파이프와 결합
ps aux | awk 'NR>1 { print $1, $3, $4, $11 }' | sort -k3 -rn | head -10
cat access.log | awk '{ print $7 }' | sort | uniq -c | sort -rn | head -20
```

---

## find

```bash
find . -name "*.go"             # 이름 패턴
find . -iname "*.Go"            # 대소문자 무시
find . -type f                  # 파일만
find . -type d                  # 디렉토리만
find . -type l                  # 심볼릭 링크
find . -maxdepth 2              # 깊이 제한
find . -mindepth 1 -maxdepth 1  # 현재 디렉토리만

# 시간 기반 (-mtime: 수정, -atime: 접근, -ctime: 변경)
find . -mtime -7                # 7일 이내 수정
find . -mtime +30               # 30일 초과
find . -newer reference.txt     # reference.txt보다 최신

# 크기
find . -size +10M               # 10MB 초과
find . -size -1k                # 1KB 미만

# 권한
find . -perm 644
find . -perm /u+x               # 실행 가능

# 실행
find . -name "*.tmp" -delete    # 삭제
find . -name "*.sh" -exec chmod +x {} \;    # 각 파일에 실행
find . -name "*.log" -exec gzip {} \;

# print0 + xargs -0 (공백 포함 파일명 안전)
find . -name "*.txt" -print0 | xargs -0 grep "pattern"
find . -mtime +30 -print0 | xargs -0 rm -f
```

---

## xargs

```bash
echo "a b c" | xargs echo           # a b c를 인자로 전달
cat files.txt | xargs rm            # 파일 목록 삭제
ls *.txt | xargs wc -l

# -I{}: 치환자 지정
cat servers.txt | xargs -I{} ssh {} "uptime"
find . -name "*.go" | xargs -I{} cp {} /backup/

# -n: 한 번에 넘길 인자 수
echo "1 2 3 4 5" | xargs -n2 echo   # 2개씩: "1 2", "3 4", "5"

# -P: 병렬 처리
cat urls.txt | xargs -P4 -I{} curl -s {}   # 4개 동시

# -0: null 구분자 (find -print0 와 조합)
find . -name "*.log" -print0 | xargs -0 -P4 gzip
```

---

## sort + uniq

```bash
sort file.txt                    # 사전순 정렬
sort -n file.txt                 # 숫자 정렬
sort -r file.txt                 # 역순
sort -k2 file.txt                # 2번째 필드 기준
sort -k2 -n -r file.txt          # 2번째 필드, 숫자 역순
sort -t: -k3 -n /etc/passwd      # : 구분, 3번째 필드

sort -u file.txt                 # 정렬 + 중복 제거

# uniq (정렬된 입력 전제)
sort file | uniq                 # 중복 제거
sort file | uniq -c              # 개수 포함
sort file | uniq -d              # 중복된 것만
sort file | uniq -u              # 유일한 것만

# 빈도 상위 N개
cat access.log | awk '{print $1}' | sort | uniq -c | sort -rn | head -20
```

---

## cut

```bash
cut -d: -f1 /etc/passwd          # : 구분, 1번째 필드 (username)
cut -d: -f1,3 /etc/passwd        # 1,3번째
cut -d, -f2-4 data.csv           # 2~4번째 필드
cut -c1-10 file.txt              # 1~10번째 문자
cut -c5- file.txt                # 5번째부터 끝

# awk가 더 유연하지만 cut은 빠름
```

---

## tr

```bash
echo "Hello" | tr 'a-z' 'A-Z'      # 소문자 → 대문자
echo "Hello" | tr 'A-Z' 'a-z'      # 대문자 → 소문자
echo "a:b:c" | tr ':' '\n'          # : → 줄바꿈
echo "foo  bar" | tr -s ' '         # 연속 공백 → 단일 공백
echo "hello123" | tr -d '0-9'       # 숫자 삭제
echo "hello" | tr -d '\n'           # 줄바꿈 삭제
cat file | tr -dc 'a-zA-Z0-9\n'    # 영숫자/줄바꿈 외 삭제
```

---

## jq

```bash
# 기본 필터
echo '{"name":"Alice","age":30}' | jq '.name'         # "Alice"
echo '{"name":"Alice","age":30}' | jq -r '.name'      # Alice (raw, 따옴표 없음)

# 배열
echo '[1,2,3]' | jq '.[]'           # 각 원소
echo '[1,2,3]' | jq '.[1]'          # 인덱스
echo '[1,2,3]' | jq '.[1:3]'        # 슬라이스

# 중첩
echo '{"a":{"b":1}}' | jq '.a.b'

# 배열 조작
jq '.[] | .name' users.json                    # 각 원소의 name
jq '[.[] | select(.age > 20)]' users.json      # 필터
jq 'map(.age)' users.json                      # 필드 추출
jq 'map(select(.active == true))' users.json

# 조합
jq '{name: .name, score: .stats.score}' data.json    # 재구조화
jq '.[] | [.name, .age] | @csv' users.json            # CSV 출력
jq 'length' arr.json                                  # 배열 길이

# curl + jq 조합
curl -s "https://api.example.com/users" | jq '.data[] | .email'

# 여러 키 추출 + 헤더
curl -s "https://api.github.com/repos/torvalds/linux" \
  | jq '{name, stars: .stargazers_count, forks}'
```

---

## tee

```bash
command | tee output.txt           # 화면 + 파일 동시 출력
command | tee -a output.txt        # 파일에 추가 (append)
command | tee file1 file2          # 여러 파일
command | tee /dev/stderr | next   # 파이프 중간 디버그 출력
```

---

## wc / diff / comm

```bash
wc -l file.txt          # 줄 수
wc -w file.txt          # 단어 수
wc -c file.txt          # 바이트 수
wc -m file.txt          # 문자 수 (멀티바이트 안전)

diff file1 file2         # 기본 diff
diff -u file1 file2      # unified format (패치 형식)
diff -r dir1 dir2        # 디렉토리 재귀 비교

comm file1 file2         # 정렬된 파일 비교
# 출력: col1=file1만, col2=file2만, col3=공통
comm -12 file1 file2     # 공통만
comm -23 file1 file2     # file1에만 있는 것
```

---

## tar + gzip

```bash
# 압축
tar -czf archive.tar.gz dir/        # gzip 압축
tar -cjf archive.tar.bz2 dir/       # bzip2 압축
tar -cJf archive.tar.xz dir/        # xz 압축
tar -cf archive.tar dir/            # 압축 없이 묶기

# 해제
tar -xzf archive.tar.gz             # 현재 디렉토리
tar -xzf archive.tar.gz -C /target  # 지정 경로
tar -xzf archive.tar.gz file.txt    # 특정 파일만

# 목록 확인
tar -tzf archive.tar.gz

# 실전: 증분 백업
tar -czf backup-$(date +%Y%m%d).tar.gz --newer-mtime="1 day ago" /data
```

---

## curl

```bash
curl https://example.com                          # GET
curl -s https://example.com                       # silent (진행 미출력)
curl -o file.html https://example.com             # 파일로 저장
curl -L https://example.com                       # 리다이렉트 따라감
curl -I https://example.com                       # 헤더만
curl -f https://example.com                       # HTTP 에러 시 exit 1

# POST
curl -X POST https://api.example.com/data \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"key":"value"}'

# JSON POST (파일)
curl -X POST https://api.example.com/data \
  -H "Content-Type: application/json" \
  -d @payload.json

# 폼 데이터
curl -X POST https://example.com/form \
  -F "name=Alice" \
  -F "file=@photo.jpg"

# 상태 코드 확인
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://example.com)
[[ "$HTTP_CODE" == "200" ]] || echo "에러: $HTTP_CODE"
```

---

## 파이프 조합 실전 패턴

```bash
# 접근 로그에서 상위 IP
awk '{print $1}' access.log | sort | uniq -c | sort -rn | head -10

# 에러 로그 시간대별 집계
grep "ERROR" app.log | awk '{print $1}' | cut -d: -f1 | sort | uniq -c

# 특정 프로세스 메모리 합계 (MB)
ps aux | grep python | grep -v grep | awk '{sum+=$6} END {printf "%.1f MB\n", sum/1024}'

# 중복 파일 찾기
find . -type f -print0 | xargs -0 md5sum | sort | uniq -D -w32

# CSV 특정 컬럼 합계
awk -F',' 'NR>1 {sum+=$3} END {print sum}' data.csv

# JSON 배열 → TSV
curl -s api/users | jq -r '.[] | [.id, .name, .email] | @tsv'

# 디렉토리 크기 상위 10
du -sh /* 2>/dev/null | sort -rh | head -10

# 로그에서 응답시간 평균
awk '/response_time/ {sum+=$NF; n++} END {print sum/n "ms"}' app.log

# 포트 사용 프로세스 확인
ss -tlnp | awk 'NR>1 {print $4, $6}' | column -t
```
