/*
 CLIENT_T
- CLIENT_UUID uuid, not null, primary key, (UUIDv7)
- CREATED_DATE timestamp with time zone, not null, default now()
- IP character varying(50) not null
- USER_AGENT character varying(100) not null
- INDEXES
  - PK: CLIENT_UUID
 */

CREATE TABLE CLIENT_T (
    CLIENT_UUID UUID NOT NULL PRIMARY KEY,
    CREATE_DATE TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    IP INET NOT NULL
)
;