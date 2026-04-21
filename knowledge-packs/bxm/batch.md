# 배치 처리 패턴

## 1. Chunk-Oriented

DB 또는 파일로부터 한 건 단위로 입력 받는 ItemReader와 한 건 단위로 비즈니스 로직을 처리하는 ItemProcessor, 결과를 출력하는 ItemWriter로 구분되는 패턴이다.

Figure 1. 다건 처리 방식(Chunk-Oriented)

단일 Step의 일반적 형태는 ItemReader, ItemProcessor, ItemWriter로 구성된다.

One-Transaction 내에서 한번씩 데이터를 Read한 후 지정된 commit-interval(chunk) 단위로 Write 하는 방식이다.

예를 들어, Commit Interval이 100으로 설정이 되어 있으면, 입력 데이터 기준으로 ItemReader와 ItemProcessor가 100건의 데이터를 처리한 후 ItemWriter로 넘어온 chunk를 DB 또는 파일에 Writer한 후 해당 트랜잭션이 Commit이 된다.

## 2. Tasklet

단순 DB CRUD 수행이나 반드시 한번에 commit/rollback되어야 하는 배치에서 사용하는 패턴이다.

Figure 2. Tasklet

단순작업(단일 DB CRUD statement 실행 등) 처리나 One 트랜잭션으로 구성되는 비즈니스 로직을 구현할 때 사용하는 Step 유형으로 execute() 메소드를 가진 Tasklet interface를 사용하여 구현한다.

배치 Job 내의 Step 안에서 chunk 단위로 트랜잭션 처리를 하는 chunk-oriented processing과는 달리 Step 실행 시 단일 트랜잭션만 수행되는 구조로 input data를 한번에 commit/rollback 처리 할 때 쓰일 수 있다.