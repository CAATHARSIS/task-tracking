ALTER TABLE users
    ALTER COLUMN password_hash TYPE VARCHAR(50),
    ALTER COLUMN email TYPE VARCHAR(50);