-- DICT LBPay - PostgreSQL Initialization Script
-- Creates databases and users for all services

-- Create databases for each service
CREATE DATABASE conn_dict;
CREATE DATABASE conn_bridge;
CREATE DATABASE core_dict;

-- Create service users with limited privileges
CREATE USER conn_dict_user WITH ENCRYPTED PASSWORD 'conn_dict_password_dev';
CREATE USER conn_bridge_user WITH ENCRYPTED PASSWORD 'conn_bridge_password_dev';
CREATE USER core_dict_user WITH ENCRYPTED PASSWORD 'core_dict_password_dev';

-- Grant privileges to service users
GRANT ALL PRIVILEGES ON DATABASE conn_dict TO conn_dict_user;
GRANT ALL PRIVILEGES ON DATABASE conn_bridge TO conn_bridge_user;
GRANT ALL PRIVILEGES ON DATABASE core_dict TO core_dict_user;

-- Connect to each database and set ownership
\c conn_dict
GRANT ALL ON SCHEMA public TO conn_dict_user;
ALTER DATABASE conn_dict OWNER TO conn_dict_user;

\c conn_bridge
GRANT ALL ON SCHEMA public TO conn_bridge_user;
ALTER DATABASE conn_bridge OWNER TO conn_bridge_user;

\c core_dict
GRANT ALL ON SCHEMA public TO core_dict_user;
ALTER DATABASE core_dict OWNER TO core_dict_user;

-- Create extensions
\c conn_dict
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

\c conn_bridge
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

\c core_dict
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Log successful initialization
\c dict
SELECT 'PostgreSQL initialization completed successfully' AS status;