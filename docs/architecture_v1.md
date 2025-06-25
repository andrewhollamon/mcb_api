# DESIGN VERSION 1

## OVERVIEW OF LAYERS

Front End: Static HTML with hand-written JS for the interactions
- Each front-end instance spawned will be given a unique UUID to act as an ID for the front-end instance

API Layer: Go-based API using in-memory queue:
- Create Frontend Instance
    - IN: IP Address, Browser String
    - OUT: Client UUID
    - POST /api/v1/client/new
    - VERY strong rate limiting, mostly IP based
- Request Checkbox
    - IN: Client UUID, Checkbox Number
    - OUT: Request UUID
    - POST /api/v1/checkbox/{checkbox_nbr}/check
    - POST /api/v1/checkbox/{checkbox_nbr}/uncheck
    - Check queue status .... if not being drained
- Request Status
    - IN: Request UUID
    - OUT: Result (Success, Failure)
    - GET /api/v1/status/{request_uuid}

Partitioning Service: Go-based internal server to handle partitioning based on checkbox number, can be reconfigured with more/less partitions

Queue: AWS queue TBD

Persistence Service:  Go-based, consume requests from the queue, and attempt persistence to the DB

DB Layer:  AWS PostgreSQL TBD

## DATABASE DESIGN

CLIENT_T
- CLIENT_UUID uuid, not null, primary key, (UUIDv7)
- CREATED_DATE timestamp with time zone, not null, default now()
- IP character varying(50) not null
- USER_AGENT character varying(100) not null
- INDEXES
  - PK: CLIENT_UUID

CHECKBOX_T  
(ranged partitioned on CHECKBOX_NBR:  https://www.postgresql.org/docs/current/ddl-partitioning.html)
- CHECKBOX_NBR int, unique, not null, primary key
- CHECKED_STATE boolean, not null, default false
- INDEXES
  - PK: CHECKBOX_NBR

CHECKBOX_DETAILS_T
- CHECKBOX_NBR int, unique, not null, primary key
- INIT_DATE timestamp with time zone, not null, default now()
- LAST_UPDATED_BY uuid, not null, (UUIDv7)
- LAST_REQUEST_ID uuid, not null, (UUIDv7)
- LAST_UPDATED_DATE timestamp with time zone, not null
- INDEXES
  - PK: CHECKBOX_NBR
  - IX1: LAST_UPDATED_BY

UPDATE_T
- UPDATE_DATE timestamp with time zone, not null, default now(), primary key
- CHECKBOX_NBR int, not null
- CHECKED boolean, not null
- UPDATED_BY uuid, not null, (UUIDv7)
- REQUEST_ID uuid, not null, (UUIDv7)
- SUCCESS boolean, not null
- INDEXES
  - PK: UPDATE_DATE
  - IX1: CHECKBOX_NBR
  - IX2: UPDATED_BY

LOG_T
- LOG_DATE timestamp with time zone, not null, default now()
- SOURCE character varying(50), not null
- TRACE_ID uuid, not null