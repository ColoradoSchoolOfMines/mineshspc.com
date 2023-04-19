-- v4: Add qrcodesent column to students table

ALTER TABLE students ADD COLUMN qrcodesent BOOLEAN NOT NULL DEFAULT FALSE;
