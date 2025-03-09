CREATE TABLE public.users
(
    id            SERIAL PRIMARY KEY,
    username      VARCHAR NOT NULL UNIQUE,
    admin         BOOLEAN NOT NULL DEFAULT FALSE,
    password_hash VARCHAR NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON public.users (username);
