# alter_db

sqlite3은 alter가 되지 않는다. 그래서 해당 table의 컬럼 타입을 변경하고 싶을 때 기존 테이블의 데이터를 옮기는 작업을 하는 코드다.
진행 중 중단할 수 있음을 전제로 하여 except를 사용하였다.
