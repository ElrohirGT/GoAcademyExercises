CREATE DATABASE exercise;
\c exercise postgres;

-- Create Tables

CREATE TABLE db_user (
    id INTEGER NOT NULL PRIMARY KEY,
    gender VARCHAR(16) NOT NULL,
    name VARCHAR(64) NOT NULL,
    location VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    state VARCHAR(255) NOT NULL,
    country VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(16) NOT NULL
);

CREATE USER root WITH PASSWORD 'root';
GRANT ALL PRIVILEGES ON DATABASE exercise TO root;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO root;

-- Creating main user the API is going to connect to...
CREATE USER backend WITH PASSWORD 'backend';
REVOKE ALL ON DATABASE exercise FROM backend;
GRANT CONNECT ON DATABASE exercise TO backend;
