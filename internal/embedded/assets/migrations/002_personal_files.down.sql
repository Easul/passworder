-- 002_personal_files.down.sql
DROP INDEX IF EXISTS idx_personal_files_name;
DROP INDEX IF EXISTS idx_personal_files_type_deleted;
DROP INDEX IF EXISTS idx_personal_files_deleted;
DROP TABLE IF EXISTS personal_files;
