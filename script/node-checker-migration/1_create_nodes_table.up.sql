CREATE TABLE nodes (
  key         varchar(255) primary key,
  last_check  timestamp not null,
  online      boolean,
  created_at  timestamp not null,
  updated_at  timestamp not null,
  deleted_at  timestamp null
);

  CREATE TABLE uptimes (
  id          serial primary key,
  node_id     varchar (255),
  start_time  integer,
  created_at  timestamp not null,
  updated_at  timestamp not null,
  deleted_at  timestamp null,
  foreign key (node_id) references nodes(key)
);