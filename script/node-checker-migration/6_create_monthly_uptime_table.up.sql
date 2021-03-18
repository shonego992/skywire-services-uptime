CREATE TABLE monthly_uptimes (
  id          serial primary key,
  node_id     varchar (255),
  month integer,
  year integer,
  total_start_time integer,
  created_at  timestamp not null,
  updated_at  timestamp not null,
  deleted_at  timestamp null,
  percentage float,
  downtime integer,
  last_start_time integer,
  foreign key (node_id) references nodes(key)
);