-- v6: Remove the retired team division fields

CREATE TABLE teams_without_division (
  id             TEXT NOT NULL,
  teacheremail   TEXT NOT NULL,
  name           TEXT NOT NULL,
  inperson       BOOLEAN NOT NULL,
  registration_ts BIGINT NOT NULL DEFAULT 0,

  PRIMARY KEY (id, teacheremail)
);

INSERT INTO teams_without_division (id, teacheremail, name, inperson, registration_ts)
SELECT id, teacheremail, name, inperson, registration_ts FROM teams;

DROP TABLE teams;
ALTER TABLE teams_without_division RENAME TO teams;
