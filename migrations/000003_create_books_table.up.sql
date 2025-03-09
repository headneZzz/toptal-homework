CREATE TABLE public.books
(
    id    SERIAL PRIMARY KEY,
    title VARCHAR NOT NULL,
    year  INT     NOT NULL CHECK (year > 0),
    author VARCHAR NOT NULL,
    price INT NOT NULL CHECK (price >= 0),
    stock INT NOT NULL CHECK (stock >= 0),
    category_id INT NOT NULL,
    CONSTRAINT fk_books_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE RESTRICT
);

CREATE INDEX idx_books_category_id ON public.books (category_id);
