--gomigrator up

create table t(i int primary key, j int);

--gomigrator down

drop table t;
