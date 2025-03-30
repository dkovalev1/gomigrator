--gomigrator up

CREATE TABLE test1(i INT PRIMARY KEY, j INT);
INSERT INTO test1(i) VALUES(1);

--gomigrator down

DROP TABLE IF EXISTS test1;
