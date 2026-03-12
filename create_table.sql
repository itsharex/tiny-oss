CREATE DATABASE IF NOT EXISTS myoss;
USE myoss;

CREATE TABLE files (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(64) UNIQUE NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    ext VARCHAR(10),
    is_private TINYINT DEFAULT 1,  -- 1=公开 2=私有
    created_at DATETIME DEFAULT NOW()
);