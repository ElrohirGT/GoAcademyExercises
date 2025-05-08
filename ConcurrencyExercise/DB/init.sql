CREATE DATABASE sanitas;
\c sanitas postgres;

-- Create Tables

CREATE TABLE User (
	Id INTEGER NOT NULL PRIMARY KEY,
	Gender VARCHAR(16) NOT NULL,
	Name VARCHAR(64) NOT NULL,
	Location VARCHAR(255) NOT NULL,
	City VARCHAR(255) NOT NULL,
	State VARCHAR(255) NOT NULL,
	Country VARCHAR(255) NOT NULL,
	Email VARCHAR(255) NOT NULL,
	Phone VARCHAR(16) NOT NULL
)

CREATE USER root WITH PASSWORD 'root';
GRANT ALL PRIVILEGES ON DATABASE sanitas TO root;

-- Creating main user the API is going to connect to...
CREATE USER backend WITH PASSWORD 'backend';
REVOKE ALL ON DATABASE sanitas FROM backend;
GRANT CONNECT ON DATABASE sanitas TO backend;
