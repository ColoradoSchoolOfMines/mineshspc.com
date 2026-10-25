-- v6: Remove retired team division and location fields

ALTER TABLE teams DROP COLUMN division;
ALTER TABLE teams DROP COLUMN divisionexplanation;
ALTER TABLE teams DROP COLUMN inperson;
