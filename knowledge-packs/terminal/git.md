# Git Commands Reference

## Repository Setup

### Initialize/Clone
```bash
git init                                    # 새 리포지토리
git clone https://github.com/user/repo.git
git clone --depth 1 url                     # shallow clone (히스토리 없이)
```

## Status & Info

### Check Status
```bash
git status
git status -s                               # 간단 모드
git log
git log --oneline
git log --oneline --graph --all
git log -n 5                                # 최근 5개 커밋
git log --since="2 weeks ago"
```

### Show Changes
```bash
git diff                                    # unstaged 변경사항
git diff --staged                           # staged 변경사항
git diff HEAD~1                             # 이전 커밋과 비교
git diff branch1..branch2
git show commit-hash                        # 특정 커밋 내용
```

## Staging & Commit

### Add Files
```bash
git add file.txt
git add .                                   # 모든 변경사항
git add -p                                  # interactive staging
git reset file.txt                          # unstage
```

### Commit
```bash
git commit -m "message"
git commit -am "message"                    # add + commit (tracked files만)
git commit --amend                          # 마지막 커밋 수정
git commit --amend --no-edit                # 메시지 유지하고 수정
```

## Branches

### Create/Switch
```bash
git branch                                  # 브랜치 목록
git branch -a                               # 모든 브랜치 (remote 포함)
git branch feature-branch                   # 브랜치 생성
git checkout feature-branch                 # 브랜치 전환
git checkout -b feature-branch              # 생성 + 전환
git switch feature-branch                   # 브랜치 전환 (새 명령)
git switch -c feature-branch                # 생성 + 전환
```

### Delete Branch
```bash
git branch -d feature-branch                # 삭제 (merged만)
git branch -D feature-branch                # 강제 삭제
git push origin --delete feature-branch     # remote 브랜치 삭제
```

## Remote

### Push/Pull/Fetch
```bash
git remote -v                               # remote 목록
git remote add origin url
git fetch origin                            # remote 업데이트 가져오기
git pull                                    # fetch + merge
git pull --rebase                           # fetch + rebase
git push
git push -u origin main                     # upstream 설정
git push --force-with-lease                 # 안전한 force push
```

## Merge & Rebase

### Merge
```bash
git merge feature-branch                    # 현재 브랜치로 merge
git merge --no-ff feature-branch            # fast-forward 방지
git merge --abort                           # merge 취소
```

### Rebase
```bash
git rebase main                             # main 위에 현재 브랜치 재배치
git rebase --continue                       # conflict 해결 후 계속
git rebase --abort                          # rebase 취소
git rebase -i HEAD~3                        # interactive rebase (최근 3개)
```

## Stash

### Save/Apply
```bash
git stash                                   # 변경사항 임시 저장
git stash push -m "message"
git stash list                              # stash 목록
git stash pop                               # 가장 최근 stash 적용 + 삭제
git stash apply stash@{0}                   # stash 적용 (유지)
git stash drop stash@{0}                    # stash 삭제
git stash clear                             # 모든 stash 삭제
```

## Reset & Revert

### Reset
```bash
git reset --soft HEAD~1                     # 커밋만 취소 (변경사항 유지, staged)
git reset --mixed HEAD~1                    # 커밋 + staging 취소 (기본)
git reset --hard HEAD~1                     # 커밋 + 변경사항 모두 삭제
git reset --hard origin/main                # remote와 동기화
```

### Revert
```bash
git revert commit-hash                      # 커밋 되돌리기 (새 커밋 생성)
git revert HEAD                             # 최근 커밋 되돌리기
```

## Cherry-pick

### Apply Specific Commit
```bash
git cherry-pick commit-hash                 # 특정 커밋만 적용
git cherry-pick branch-name                 # 브랜치의 최신 커밋 적용
```

## Tags

### Create/Push Tags
```bash
git tag v1.0.0                              # 태그 생성
git tag -a v1.0.0 -m "Release 1.0"          # annotated 태그
git tag                                     # 태그 목록
git push origin v1.0.0                      # 태그 push
git push origin --tags                      # 모든 태그 push
git tag -d v1.0.0                           # 로컬 태그 삭제
git push origin --delete v1.0.0             # remote 태그 삭제
```

## 자주 쓰는 패턴

### 실수한 커밋 수정
```bash
# 마지막 커밋에 파일 추가
git add forgotten-file.txt
git commit --amend --no-edit

# 커밋 메시지만 수정
git commit --amend -m "new message"
```

### Conflict 해결
```bash
git merge feature-branch
# conflict 발생 시 파일 수정
git add resolved-file.txt
git commit
```

### 과거 커밋으로 돌아가기
```bash
git log --oneline
git checkout commit-hash                    # detached HEAD 상태
git checkout -b new-branch                  # 새 브랜치로 만들기
```
