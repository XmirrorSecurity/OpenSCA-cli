--
-- File generated with SQLiteStudio v3.4.3 on 22:18:37 2023-07-28
--
-- Text encoding used: System
--
PRAGMA foreign_keys = off;
BEGIN TRANSACTION;

-- Table: component
CREATE TABLE IF NOT EXISTS component (
    id       INTEGER       PRIMARY KEY AUTOINCREMENT,
    name     VARCHAR (50)  NOT NULL,
    version  VARCHAR (50)  NOT NULL,
    vendor   VARCHAR (50) 
                           DEFAULT 'N.A.',
    language VARCHAR (50)  NOT NULL,
    purl     VARCHAR (256) NOT NULL
                           UNIQUE
);


-- Table: reference
CREATE TABLE IF NOT EXISTS reference (
    id          INTEGER       PRIMARY KEY AUTOINCREMENT,
    module_name VARCHAR (100) NOT NULL,
    purl        VARCHAR (256) NOT NULL
);


-- Index: uk_ref
CREATE UNIQUE INDEX IF NOT EXISTS uk_ref ON reference (
    module_name,
    purl
);


COMMIT TRANSACTION;
PRAGMA foreign_keys = on;
