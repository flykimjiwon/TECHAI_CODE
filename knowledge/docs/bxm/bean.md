# Bean 작성

## 1. 비즈니스 로직 작성

Bean 클래스를 생성한다. 패키지 익스플로러에서 우클릭 → New → 새로운 Bean을 선택한다.

Figure 1. Bean 생성

새로운 BXM Bean Wizard에서 타입 이름과 논리 이름을 입력한다.

Figure 2. 새로운 BXM Bean Wizard

비즈니스 로직을 작성한다. BXM 개발표준은 BXM 개발자가이드 - 개발표준 문서를 참고한다. 아래는 작성한 Bean 메소드이다.

```
@BxmBean
@BxmCategory(logicalName = "Employee Info Management")
public class MSmpEmpInfMng {
    final Logger logger = LoggerFactory.getLogger(this.getClass());

    private DSmpEmpTst000 dSmpEmpTst000;

    /**
     * Select a single employee info.
     *
     * @param   input   DSmpEmpTst000Dto
     * @return DSmpEmpTst000Dto
     * @throws DefaultApplicationException
     */
    @BxmCategory(logicalName = "Single Select")
    public DSmpEmpTst000Dto getEmpInf(DSmpEmpTst000Dto input) throws DefaultApplicationException {

        logger.debug("============== START ==============");
        logger.debug("input = {}", input);

        dSmpEmpTst000 = DefaultApplicationContext.getBean(dSmpEmpTst000, DSmpEmpTst000.class);

        /**
         * @BXMType VariableDeclaration
         */
        DSmpEmpTst000Dto output = null;

        /**
         * @BXMType DbioCall
         * Employee ID number selectOne
         */
        output = dSmpEmpTst000.selectOne00(input);

        logger.debug("output = {}", output);
        logger.debug("============== END ==============");

        return output;
    }
}
```

## 2. 모듈 테스트 생성

모듈 테스트를 생성한다. 패키지 익스플로러에서 우클릭 → New → 새로운 모듈 테스트를 선택한다.

Figure 3. 모듈 테스트 생성

빈 클래스를 선택하지 않은 경우 클래스를 검색하여 입력한다.

검색 결과 중 원하는 메소드를 선택한 후 Finish버튼을 누른다.

Figure 4. 모듈 테스트 생성

아래는 생성된 모듈 테스터이며 테스트는 서버 배포 후 실행한다. 모듈 테스터 실행은 모듈 테스트 수행를 참고한다.

Figure 5. 모듈 테스터 화면