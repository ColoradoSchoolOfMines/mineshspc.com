-- v2: Add registration timestamp to the teams table

ALTER TABLE teams ADD COLUMN registration_ts BIGINT NOT NULL DEFAULT 0;
