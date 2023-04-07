-- v1: Initial tables

CREATE TABLE IF NOT EXISTS teachers (
  email          TEXT NOT NULL PRIMARY KEY,
  name           TEXT NOT NULL,
  emailconfirmed BOOLEAN NOT NULL DEFAULT FALSE,
  emailallowance INTEGER NOT NULL DEFAULT 16,

  -- school info is 1:1 with teachers
  schoolname     TEXT,
  schoolcity     TEXT,
  schoolstate    TEXT
);

CREATE TABLE IF NOT EXISTS teams (
  id                  TEXT NOT NULL,
  teacheremail        TEXT NOT NULL,
  name                TEXT NOT NULL,
  division            TEXT NOT NULL,
  divisionexplanation TEXT NOT NULL,
  inperson            BOOLEAN NOT NULL,

  PRIMARY KEY (id, teacheremail)
);


CREATE TABLE IF NOT EXISTS students (
  email                  TEXT NOT NULL PRIMARY KEY,
  teamid                 TEXT NOT NULL,
  name                   TEXT NOT NULL,
  age                    INTEGER NOT NULL,
  parentemail            TEXT,
  signatory              TEXT,
  dietaryrestrictions    TEXT,
  campustour             BOOLEAN,
  previouslyparticipated BOOLEAN NOT NULL DEFAULT FALSE,
  emailconfirmed         BOOLEAN NOT NULL DEFAULT FALSE,

  -- Waivers
  liabilitywaiver        BOOLEAN NOT NULL DEFAULT FALSE,
  computerusewaiver      BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS admins (
  email TEXT NOT NULL PRIMARY KEY
);
