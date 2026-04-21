# 개발 프로세스

온라인 어플리케이션의 개발 프로세스는 일반적으로 다음과 같다.

Figure 1. 개발 프로세스

## 1. 온라인 어플리케이션 샘플 클래스 및 메소드 정의

본 장에서 단건 조회 서비스를 만들 때 사용하는 클래스/메소드 명이다. 샘플을 작성할 때 동일한 이름을 사용하지 않아도 되며, 가이드 이해를 돕기 위해 참고한다.

샘플 클래스 및 메소드 명

| 컴포넌트 종류 
| 컴포넌트 명 
| 비고 

어플리케이션 명(프로젝트 명)

BxmDftSmp

|  

Service 클래스 명

SSMP1001A

|  

Service Operation 명

ssmp1001a001

|  

Bean 클래스 명

MSmpEmpInfMng

|  

Bean 메소드 명

getEmpInf

|  

DBIO 명

DSmpEmpTst000

|  

SQL ID

selectOne00

|  

IO 명 (Service 사용)

SSMP1001A001InDto, SSMP1001A001OutDto

Service에 사용하는 입/출력 IO이다.

IO 명 (Bean, DBIO 사용)

DSmpEmpTst000Dto

데이터에 접근하거나 전달할 때 사용하는 IO이다.