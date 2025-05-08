CREATE DATABASE sanitas;
\c sanitas postgres;

-- Create Tables


CREATE USER root WITH PASSWORD 'root';
GRANT ALL PRIVILEGES ON DATABASE sanitas TO root;

-- Creating main user the API is going to connect to...
CREATE USER backend WITH PASSWORD 'backend';
REVOKE ALL ON DATABASE sanitas FROM backend;
GRANT CONNECT ON DATABASE sanitas TO backend;
