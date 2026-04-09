# Service 작성

## 1. Service IO 작성

Service IO 작성도 DBIO IO 작성과 동일하다. IO 생성 및 작성 방법은 DBIO IO 생성를 참고한다. 아래는 생성된 IO 화면이다.

Service IO는 bxm.dft.smp.onlne.service 하위에 dto 패키지 하위에 작성했다.

Figure 1. 단건조회 Input : SSMP1001A001InDto

Figure 2. 단건조회 Output : SSMP1001A001OutDto

## 2. Service 작성

Service 클래스를 생성한다. 패키지 익스플로러에서 우클릭 → New → 새로운 서비스를 선택한다.

Figure 3. Service 생성

새로운 BXM Service Wizard에서 타입 이름과 논리 이름을 입력한다.

Figure 4. 새로운 Service Wizard

로직을 작성한다. BXM 개발표준은 BXM 개발자가이드 - 개발표준 문서를 참고한다. 아래는 작성한 Service 클래스와 Service Operation이다.

```
@BxmService("SSMP1001A")
@BxmCategory(logicalName = "Single select")
public class SSMP1001A {
    final Logger logger = LoggerFactory.getLogger(this.getClass());

    private MSmpEmpInfMng mSmpEmpInfMng;

    @BxmServiceOperation("ssmp1001a001")
    @BxmCategory(logicalName = "Single select")
    public SSMP1001A001OutDto ssmp1001a001(SSMP1001A001InDto input) throws DefaultApplicationException {

        logger.debug("============== SERVICE START ==============");
        logger.debug("input = {}", input);

        mSmpEmpInfMng = DefaultApplicationContext.getBean(mSmpEmpInfMng, MSmpEmpInfMng.class);

        /**
         * @BXMType VariableDeclaration
         */
        DSmpEmpTst000Dto beanInput = new DSmpEmpTst000Dto();

        /**
         * @BXMType VariableDeclaration
         */
        SSMP1001A001OutDto output = new SSMP1001A001OutDto();

        /**
         * @BXMType IF
         */
        if (input.getFeduEmpNo().equals(BigDecimal.valueOf(9877))) {
            logger.error("Pre-Deploy Test Exception for FeduEmpNo [9877].");
            throw new DefaultApplicationException("BXME30000", new Object[] {},
                    new Object[] { "Pre-Deploy Test Exception." });
        }

        /**
         * @BXMType LogicalArea
         * @Desc DTO mapping
         */
        {
            beanInput.setFeduEmpNo(input.getFeduEmpNo());
        }

        /**
         * @BXMType BeanCall
         * @Desc Call bean single select method
         */
        DSmpEmpTst000Dto beanOutput = mSmpEmpInfMng.getEmpInf(beanInput);

        /**
         * @BXMType IF
         * @Desc DTO mapping if bean out is not null
         */
        if (beanOutput != null) {
            output.setFeduEmpNo(beanOutput.getFeduEmpNo());
            output.setFeduEmpNm(beanOutput.getFeduEmpNm());
            output.setFeduOccpNm(beanOutput.getFeduOccpNm());
            output.setFeduMngrEmpNo(beanOutput.getFeduMngrEmpNo());
            output.setFeduIpsaDt(beanOutput.getFeduIpsaDt());
            output.setFeduPayAmt(beanOutput.getFeduPayAmt());
            output.setFeduDeptNo(beanOutput.getFeduDeptNo());
        }

        /**
         * @BXMType LogicalArea
         * @Desc add message
         */
        {
            DefaultApplicationContext.addMessage("BXMI60000", null, new Object[] {});
        }

        logger.debug("output = {}", output);
        logger.debug("============== SERVICE END ==============");

        return output;
    }
}
```

## 3. 새로운 서비스 테스트 생성

서비스 테스트를 생성한다. 패키지 익스플로러에서 우클릭 → New → 새로운 서비스 테스트를 선택한다.

Figure 5. 서비스 테스트 생성

오퍼레이션을 선택 후 Finish를 클릭한다.

Figure 6. 서비스 오퍼레이션 선택

아래는 생성된 서비스 테스트이며 테스트는 서버 배포 후 실행한다. 서비스 테스트의 실행은 서비스 테스트 수행를 참고한다.

Figure 7. 서비스 테스트 화면