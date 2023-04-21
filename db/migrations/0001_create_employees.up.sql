CREATE TABLE IF NOT EXISTS employees (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    salary DECIMAL(19, 2) NOT NULL,
    department VARCHAR(50) NOT NULL,
    UNIQUE KEY `first_name_last_name` (`first_name`,`last_name`)
) WITH SYSTEM VERSIONING;
