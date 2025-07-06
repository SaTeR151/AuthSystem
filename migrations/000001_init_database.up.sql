CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE IF NOT EXISTS users_auth(
    id uuid DEFAULT uuid_generate_v4 (),
    user_id uuid UNIQUE NOT NULL,
    refresh_t TEXT,
    user_agent TEXT,
    user_ip TEXT,
    PRIMARY KEY (id)
);
INSERT INTO users_auth (user_id) VALUES ('090bb747-d6d3-4067-a1da-2b83726eb24d');
INSERT INTO users_auth (user_id) VALUES ('2df8716b-d385-4b7e-aae9-4618996c438a');