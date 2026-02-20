-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

-- 1. Remove index for username (no longer needed)
DROP INDEX IF EXISTS idx_users_username;

-- 2. Remove username column
ALTER TABLE users
DROP COLUMN IF EXISTS username;