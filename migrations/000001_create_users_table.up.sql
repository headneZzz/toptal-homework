-- Создание таблицы users
CREATE TABLE public.users
(
    id            SERIAL PRIMARY KEY,
    username      VARCHAR NOT NULL UNIQUE,
    admin         BOOLEAN NOT NULL DEFAULT FALSE,
    password_hash VARCHAR NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для ускорения поиска по имени пользователя
CREATE INDEX idx_users_username ON public.users (username);
