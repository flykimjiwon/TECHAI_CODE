# 다건 Select

일반적인 Select 문으로 조회 결과가 0~N건이 된다. N값이 온라인 서버에 설정된 값을 넘으면 에러가 발생한다.

## 1. 형식

dbio의 출력 타입을 지정하면서 [List]를 체크한다.

Figure 1. 다건 Select 형식

## 2. 개발 및 유의사항

목록 조회의 결과는 java.lang.List 타입으로 리턴한다. 데이터가 없는 경우 리스트의 사이즈는 0이다.

## 3. DBIO 샘플

Figure 2. 다건 Select DBIO 화면

입력/출력 IO는 단건 조회 서비스 작성에 썼던 DSmpEmpTst000Dto IO를 재사용했다. 아래는 샘플에 사용한 SQL 문이다.

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
 <where>
	 <if test="feduEmpNm !=null and feduEmpNm !=''">
	 	AND A.FEDU_EMP_NM LIKE  #{feduEmpNm} || '%'
	 </if>
 </where>
 ORDER BY A.FEDU_EMP_NM
```

## 4. Bean 샘플

```
@BxmCategory(logicalName = "Single Select")
public DSmpEmpTst000Dto getEmpInf(DSmpEmpTst000Dto input) throws DefaultApplicationException {

		logger.debug("============== START ==============");
		logger.debug("input = {}", input);

		// Dbio 생성
		dSmpEmpTst001 = DefaultApplicationContext.getBean(dSmpEmpTst001, DSmpEmpTst001.class);

		List<DSmpEmpTst000Dto> output = null;

		// Dbio call
		output = dSmpEmpTst001.selectList01(input);

		if(output.isEmpty()) {
			//조회된 데이터가 없는 경우 처리
		}

		for(DSmpEmpTst000Dto data : output) {
			//조회된 데이터 처리
		}

		logger.debug("output = {}", output);
		logger.debug("============== END ==============");

		return output;
	}
}
```