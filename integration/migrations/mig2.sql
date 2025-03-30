--gomigrator up

CREATE TABLE test2("key" VARCHAR PRIMARY KEY, j INT);
INSERT INTO test2("key", j) VALUES('one', 1);

--gomigrator down

DROP TABLE IF EXISTS test2;
