-- https://dba.stackexchange.com/questions/117109/how-to-manage-default-privileges-for-users-on-a-database-vs-schema/117661#117661
CREATE DATABASE test_db OWNER test_user;
REVOKE ALL ON DATABASE test_db FROM public;

GRANT CONNECT ON DATABASE test_db TO test_user;

\connect test_db

CREATE SCHEMA test_db_schema AUTHORIZATION test_user;

SET search_path = test_db_schema;

ALTER ROLE test_user IN DATABASE test_db SET search_path = test_db_schema;

GRANT USAGE  ON SCHEMA test_db_schema TO test_user;
GRANT CREATE ON SCHEMA test_db_schema TO test_user;

ALTER DEFAULT PRIVILEGES FOR ROLE test_user
GRANT SELECT ON TABLES TO test_user;

ALTER DEFAULT PRIVILEGES FOR ROLE test_user
GRANT INSERT, UPDATE, DELETE ON TABLES TO test_user;
