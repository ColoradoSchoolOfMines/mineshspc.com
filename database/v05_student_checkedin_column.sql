-- v5: Add checkedin column to students table

ALTER TABLE students ADD COLUMN checkedin BOOLEAN NOT NULL DEFAULT FALSE;
