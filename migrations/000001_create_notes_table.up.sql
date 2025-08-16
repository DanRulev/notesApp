CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE CHECK (username <> ''),
    email VARCHAR(255) NOT NULL UNIQUE CHECK (email <> ''),
    password VARCHAR(255) NOT NULL CHECK (password <> ''),
    image_url VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tokens(
    user_id UUID NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
    token_id VARCHAR(255) NOT NULL CHECK (token_id <> ''),
    expired_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, token_id) 
);

CREATE TABLE IF NOT EXISTS notes(
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,  
    heading VARCHAR(255) NOT NULL CHECK (heading <> ''),
    content TEXT NOT NULL CHECK (content <> ''),
    done BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);