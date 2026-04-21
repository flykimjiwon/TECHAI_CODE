# DBIO 작성

## 1. 기능

업무에서 사용하는 SQL을 개발하는 리소스이며, 데이터 컴포넌트의 물리적인 형태이다.

## 2. DBIO IO 생성

DBIO에 사용할 IO를 생성한다. 패키지 익스플로러에서 우클릭 → New → 새로운 IO를 선택한다. (단축키 : Ctrl + N →BXM→새로운 IO 선택)

Figure 1. DBIO IO 생성

IO 이름을 입력한다.

DBIO IO는 bxm.dft.smp.onlne.dbio 하위에 dto 패키지 하위에 작성했다.

Figure 2. IO Wizard

생성된 IO에 필드를 정의한다. DB→IO 버튼을 클릭하면 아래와 같은 팝업이 뜨며, 원하는 테이블의 컬럼을 필드로 가져올 수 있다. 또는 직접 입력한다.

Figure 3. DB스키마로부터 IO 필드 생성

아래는 생성된 IO이다. rowId는 사용을 위해 임의로 추가하였다.

Figure 4. 생성된 DSmpEmpTst000Dto IO

## 3. DBIO 작성

DBIO를 생성한다. 패키지 익스플로러에서 우클릭 → New → 새로운 DBIO를 선택한다.

Figure 5. DBIO 생성

DBIO 이름과 논리 이름을 작성한다.

Figure 6. 새로운 DBIO Wizard

생성된 DBIO에 SQL ID를 입력 후 + 버튼을 눌러 추가한다.

Figure 7. SQL ID 목록

SQL과 입력/출력 타입을 입력한다. 자세한 DBIO 작성법은 *BXM 사용자가이드 - Studio*를 참고한다. 본 가이드에서는 DBIO 작성법을 간단하게 설명한다.

1 : 기본 SQL 버튼을 클릭하면 선택한 테이블과 Statement 종류에 따라 자동으로 SQL이 생성된다.

2 : 자동으로 생성된 SQL을 볼 수 있으며, 직접 수정/입력이 가능하다.

3 : 입력 타입을 검색해서 추가한다.

4 : 출력 타입을 검색해서 추가한다.

Figure 8. DBIO 작성

아래는 입력한 SQL이다.

```
SELECT
       A.FEDU_EMP_NO AS feduEmpNo
     , A.FEDU_EMP_NM AS feduEmpNm
     , A.FEDU_OCCP_NM AS feduOccpNm
     , A.FEDU_MNGR_EMP_NO AS feduMngrEmpNo
     , A.FEDU_HIRE_DT AS feduHireDt
     , A.FEDU_PAY_AMT AS feduPayAmt
     , A.FEDU_DEPT_NO AS feduDeptNo
FROM SMP_EMP_TST A
WHERE A.FEDU_EMP_NO = #{feduEmpNo}
```

## 4. SQL 테스트

SQL 테스트는 입력 영역에 테스트 값을 입력 후, SQL 테스트 버튼을 누르면 실행된다.

Figure 9. 테스트 값 입력 및 SQL 테스트 버튼

아래는 정상 실행된 SQL 테스트 결과 화면이다.

Figure 10. SQL 테스트 결과화면