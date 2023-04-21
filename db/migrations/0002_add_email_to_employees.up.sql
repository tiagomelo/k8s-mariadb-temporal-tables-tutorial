BEGIN;
SET @@system_versioning_alter_history = 1;
ALTER TABLE employees
    ADD COLUMN email varchar(100) UNIQUE NOT NULL AFTER department;
COMMIT;