CREATE ROLE MCBADMINROLE PASSWORD 'MCBADMINROLE';
CREATE ROLE MCBUSERROLE PASSWORD 'MCBUSERROLE';

CREATE USER MCBADMINUSER WITH LOGIN PASSWORD 'MCBADMINUSER';
CREATE USER MCBUSER WITH LOGIN PASSWORD 'MCBUSER';

GRANT MCBADMINROLE TO MCBADMINUSER;
GRANT MCBUSERROLE TO MCBUSER;

CREATE DATABASE MILLCHECKDB WITH OWNER = MCBADMINROLE;
GRANT CONNECT, CREATE ON DATABASE MILLCHECKDB TO MCBADMINROLE;
REVOKE ALL ON DATABASE MILLCHECKDB FROM PUBLIC;
REVOKE ALL ON DATABASE MILLCHECKDB FROM POSTGRES;
