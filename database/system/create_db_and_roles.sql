CREATE ROLE MCBADMINROLE PASSWORD 'mcbadminrole';
CREATE ROLE MCBUSERROLE PASSWORD 'mcbuserrole';

CREATE USER MCBADMINUSER WITH LOGIN PASSWORD 'mcbadminuser';
CREATE USER MCBUSER WITH LOGIN PASSWORD 'mcbuser';

GRANT MCBADMINROLE TO MCBADMINUSER;
GRANT MCBUSERROLE TO MCBUSER;

CREATE DATABASE MILLCHECKDB WITH OWNER = MCBADMINROLE;
GRANT CONNECT, CREATE ON DATABASE MILLCHECKDB TO MCBADMINROLE;
REVOKE ALL ON DATABASE MILLCHECKDB FROM PUBLIC;
REVOKE ALL ON DATABASE MILLCHECKDB FROM POSTGRES;
