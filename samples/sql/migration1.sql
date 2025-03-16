-- gomigratorUp 1

create table t(int i primary key, int j);

-- gomigratorDown 1

drop table t;
