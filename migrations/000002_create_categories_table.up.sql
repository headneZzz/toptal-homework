CREATE TABLE public.categories
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE
);
