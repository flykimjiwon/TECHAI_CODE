# Bash Scripting Reference

## Shebang

```bash
#!/bin/bash                  # 절대 경로 — /bin/bash 없으면 실패
#!/usr/bin/env bash          # PATH에서 bash 탐색 — 이식성 높음 (macOS, nix 환경 권장)
#!/bin/sh                    # POSIX sh — bash 전용 기능 사용 불가
```

## set 옵션

```bash
set -e          # 명령 실패 시 즉시 종료 (exit on error)
set -u          # 미정의 변수 사용 시 에러
set -o pipefail # 파이프 중간 실패도 감지 (set -e만으로는 파이프 실패 미감지)
set -x          # 실행 전 명령 출력 (디버그)

# 실전: 스크립트 상단에 항상
set -euo pipefail
```

pipefail 없으면: `false | true` → exit code 0 (버그 숨김)  
pipefail 있으면: `false | true` → exit code 1

---

## 변수

```bash
NAME="hello"                 # 할당 (= 양쪽 공백 없음)
readonly PI=3.14             # 상수 (재할당 불가)
export PATH="$PATH:/opt/bin" # 자식 프로세스에 전달
local x=5                    # 함수 내부 전용 (함수 밖에서 사용 불가)

# 기본값 패턴
${VAR:-default}   # VAR 미설정/빈값이면 default 반환 (VAR 변경 없음)
${VAR:=default}   # VAR 미설정/빈값이면 default로 설정 후 반환
${VAR:+other}     # VAR 설정되어 있으면 other 반환, 아니면 빈값
${VAR:?에러메시지} # VAR 미설정이면 에러 출력 후 종료

# 예시
DB_HOST="${DB_HOST:-localhost}"
PORT="${PORT:=8080}"
: "${API_KEY:?API_KEY must be set}"  # 필수값 검증
```

---

## 문자열 조작

```bash
STR="Hello, World!"

# 길이
echo ${#STR}               # 13

# 슬라이스 ${var:offset:length}
echo ${STR:0:5}            # Hello
echo ${STR:7}              # World!
echo ${STR: -6}            # orld!  (음수: 끝에서)

# 치환
echo ${STR/World/Bash}     # Hello, Bash!   (첫 번째만)
echo ${STR//l/L}           # HeLLo, WorLd!  (전부)

# 대소문자 (Bash 4+)
echo ${STR,,}              # hello, world!  (전체 소문자)
echo ${STR^^}              # HELLO, WORLD!  (전체 대문자)
echo ${STR,}               # hELLO, WORLD!  (첫 글자만 소문자)

# 접두/접미 제거
FILE="report.tar.gz"
echo ${FILE%.gz}           # report.tar      (짧은 접미 제거)
echo ${FILE%%.*}           # report          (긴 접미 제거)
echo ${FILE#*/}            # 짧은 접두 제거
echo ${FILE##*/}           # basename 효과
```

---

## 조건문

```bash
# if/elif/else
if [[ "$STATUS" == "ok" ]]; then
  echo "정상"
elif [[ "$STATUS" == "warn" ]]; then
  echo "경고"
else
  echo "에러"
fi

# [[ ]] vs [ ]
# [[ ]] — Bash 전용, 더 안전 (word splitting 없음, 패턴 매칭, && || 지원)
# [ ]   — POSIX sh, 변수 쿼팅 필수
[[ -n "$VAR" && "$VAR" != "skip" ]]   # OK
[ -n "$VAR" -a "$VAR" != "skip" ]     # POSIX 방식

# 비교 연산자
# 숫자
[[ $A -eq $B ]]   # equal
[[ $A -ne $B ]]   # not equal
[[ $A -lt $B ]]   # less than
[[ $A -gt $B ]]   # greater than
[[ $A -le $B ]]   # less or equal
[[ $A -ge $B ]]   # greater or equal

# 문자열
[[ "$A" == "$B" ]]
[[ "$A" != "$B" ]]
[[ "$A" < "$B" ]]    # 사전순 비교
[[ "$A" =~ ^[0-9]+$ ]]  # 정규식 매칭 ([[ ]] 전용)

# 빈값 체크
[[ -z "$VAR" ]]   # 빈 문자열이면 true
[[ -n "$VAR" ]]   # 비어있지 않으면 true
```

---

## 파일 테스트

```bash
[[ -e "$F" ]]   # 존재
[[ -f "$F" ]]   # 일반 파일
[[ -d "$F" ]]   # 디렉토리
[[ -r "$F" ]]   # 읽기 가능
[[ -w "$F" ]]   # 쓰기 가능
[[ -x "$F" ]]   # 실행 가능
[[ -s "$F" ]]   # 크기 > 0
[[ -L "$F" ]]   # 심볼릭 링크
[[ -h "$F" ]]   # 심볼릭 링크 (동일)
[[ "$A" -nt "$B" ]]  # A가 B보다 최신
[[ "$A" -ot "$B" ]]  # A가 B보다 오래됨

if [[ ! -f "/etc/config.yml" ]]; then
  echo "설정 파일 없음" >&2
  exit 1
fi
```

---

## 반복문

```bash
# for in (배열/목록)
for item in a b c; do
  echo "$item"
done

# for in (글로브)
for f in /var/log/*.log; do
  echo "처리: $f"
done

# C-style for (Bash)
for ((i=0; i<10; i++)); do
  echo "$i"
done

# while
COUNT=0
while [[ $COUNT -lt 5 ]]; do
  echo "$COUNT"
  ((COUNT++))
done

# until (조건이 참이 될 때까지)
until ping -c1 google.com &>/dev/null; do
  echo "네트워크 대기..."
  sleep 2
done

# 파일 라인별 읽기
while IFS= read -r line; do
  echo ">> $line"
done < /etc/hosts

# 명령 출력 라인별
while IFS= read -r line; do
  echo "$line"
done < <(find . -name "*.sh")
```

---

## 배열

```bash
# 선언 및 초기화
ARR=("apple" "banana" "cherry")
ARR[3]="date"

# 접근
echo "${ARR[0]}"        # apple
echo "${ARR[@]}"        # 모든 원소 (공백 포함 안전)
echo "${ARR[*]}"        # 모든 원소 (하나의 문자열)
echo "${#ARR[@]}"       # 원소 개수
echo "${!ARR[@]}"       # 인덱스 목록

# 슬라이스 ${arr[@]:offset:length}
echo "${ARR[@]:1:2}"    # banana cherry

# 추가
ARR+=("elderberry")

# 순회
for item in "${ARR[@]}"; do
  echo "$item"
done

# 연관 배열 (Bash 4+)
declare -A MAP
MAP["key1"]="value1"
MAP["key2"]="value2"
echo "${MAP["key1"]}"
for key in "${!MAP[@]}"; do
  echo "$key => ${MAP[$key]}"
done
```

---

## 함수

```bash
# 선언
my_func() {
  local name="$1"          # 인자: $1, $2, ..., $@
  local count="${2:-10}"   # 기본값
  echo "Hello, $name (count: $count)"
}

# 호출
my_func "World"
my_func "Bash" 5

# return vs echo
# return: exit code만 (0-255)
# echo:   문자열 반환 — $() 캡처로 사용

get_value() {
  local result="computed"
  echo "$result"           # 반환
}
VALUE=$(get_value)         # 캡처

# 에러 반환 패턴
validate() {
  [[ -n "$1" ]] || { echo "인자 필요" >&2; return 1; }
  return 0
}
validate "" || exit 1
```

---

## trap — 정리 패턴

```bash
# trap 'handler' SIGNAL
trap cleanup EXIT           # 스크립트 종료 시 (정상/비정상 모두)
trap cleanup INT TERM       # Ctrl+C, kill 시

TMPFILE=""

cleanup() {
  [[ -n "$TMPFILE" && -f "$TMPFILE" ]] && rm -f "$TMPFILE"
  echo "정리 완료" >&2
}

trap cleanup EXIT ERR INT TERM

TMPFILE=$(mktemp)
# ... 작업 ...

# ERR trap (set -e와 함께)
on_error() {
  echo "에러 발생: 라인 ${BASH_LINENO[0]}" >&2
}
trap on_error ERR
```

---

## Here-doc

```bash
# 변수 치환 O
cat <<EOF
이름: $NAME
호스트: $(hostname)
EOF

# 변수 치환 X (싱글쿼트)
cat <<'EOF'
이름: $NAME     <- 그대로 출력
호스트: $(hostname)
EOF

# 들여쓰기 제거 (탭만)
cat <<-EOF
	첫 줄 (탭 들여쓰기)
	둘째 줄
	EOF

# 변수에 저장
SQL=$(cat <<EOF
SELECT *
FROM users
WHERE id = $USER_ID;
EOF
)
```

---

## 서브셸 & 프로세스 치환

```bash
# $() — 명령 치환 (권장)
DATE=$(date +%Y%m%d)
FILES=$(find . -name "*.go" | wc -l)

# `` — 구식 백틱 (중첩 어려움)
DATE=`date +%Y%m%d`

# 프로세스 치환 <() — 명령 출력을 파일처럼
diff <(sort file1.txt) <(sort file2.txt)
while read line; do ...; done < <(command)

# 서브셸 (현재 셸 환경 영향 없음)
(cd /tmp && ls)   # 현재 디렉토리 불변
```

---

## Exit Code & 에러 처리

```bash
echo "ok"           # exit code: 0 (성공)
false               # exit code: 1
ls /없는경로        # exit code: 2

echo "exit: $?"     # 직전 명령 exit code

# 조건 실행
mkdir dir && cd dir         # 성공 시에만 다음 실행
cat file || echo "없음"     # 실패 시 대안 실행

# set -e 주의: 명령을 if 조건에 쓰면 실패해도 종료 안 함
if grep "pattern" file; then   # grep 실패해도 set -e 무시
  echo "found"
fi
```

---

## getopts — 옵션 파싱

```bash
usage() {
  echo "Usage: $0 [-h] [-v] [-f file] [-n count]"
  exit 1
}

VERBOSE=false
FILE=""
COUNT=1

while getopts ":hvf:n:" opt; do
  case $opt in
    h) usage ;;
    v) VERBOSE=true ;;
    f) FILE="$OPTARG" ;;
    n) COUNT="$OPTARG" ;;
    :) echo "옵션 -$OPTARG 에 인자 필요" >&2; exit 1 ;;
    \?) echo "알 수 없는 옵션: -$OPTARG" >&2; exit 1 ;;
  esac
done
shift $((OPTIND - 1))   # 파싱 후 남은 인자로 이동
# $@ 에 나머지 인자
```

---

## 실전 스크립트 템플릿

```bash
#!/usr/bin/env bash
set -euo pipefail

# ── 상수 ──────────────────────────────────────────────
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_NAME="$(basename "$0")"
readonly LOG_FILE="/var/log/${SCRIPT_NAME%.sh}.log"

# ── 로깅 ──────────────────────────────────────────────
log()  { echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO]  $*" | tee -a "$LOG_FILE"; }
warn() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] [WARN]  $*" | tee -a "$LOG_FILE" >&2; }
err()  { echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $*" | tee -a "$LOG_FILE" >&2; }
die()  { err "$*"; exit 1; }

# ── 정리 ──────────────────────────────────────────────
TMPDIR_WORK=""
cleanup() {
  [[ -n "$TMPDIR_WORK" ]] && rm -rf "$TMPDIR_WORK"
  log "스크립트 종료"
}
trap cleanup EXIT

# ── 메인 ──────────────────────────────────────────────
main() {
  : "${REQUIRED_VAR:?REQUIRED_VAR must be set}"

  TMPDIR_WORK=$(mktemp -d)
  log "시작: TMPDIR=$TMPDIR_WORK"

  # 작업
  log "완료"
}

main "$@"
```
