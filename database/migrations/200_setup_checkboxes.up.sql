-- create all the checkboxes, default to false (unchecked)
INSERT INTO MCB.CHECKBOX_T ( CHECKBOX_NBR, CHECKED_STATE )
SELECT a.n, FALSE
FROM GENERATE_SERIES(0,999999) AS A(N)
;

-- one record in checkbox_details_t per record in checkbox_t, with default starting values
do $$
    declare nowts timestamp;
    declare niluuid uuid;
begin
        nowts := timezone('America/Phoenix', now());
        niluuid := '00000000-0000-0000-0000-000000000000';
INSERT INTO MCB.CHECKBOX_DETAILS_T ( CHECKBOX_NBR, INIT_DATE, LAST_UPDATED_BY, LAST_REQUEST_ID, LAST_UPDATED_DATE )
SELECT a.n,
       nowts,
       niluuid,
       niluuid,
       nowts
FROM GENERATE_SERIES(0,999999) AS A(N)
;
end;
$$
;
