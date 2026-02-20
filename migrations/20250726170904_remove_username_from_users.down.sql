-- +migrate Down
-- SQL in section 'Down' is executed when this migration is rolled back

-- Since the table was empty, just add the column back
ALTER TABLE users
ADD COLUMN username VARCHAR(50) NOT NULL UNIQUE;