-- Создание таблицы cart
CREATE TABLE public.cart
(
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL,
    book_id    INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_cart_user FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE CASCADE,
    CONSTRAINT fk_cart_book FOREIGN KEY (book_id) REFERENCES public.books (id) ON DELETE CASCADE,
    CONSTRAINT unique_user_book UNIQUE (user_id, book_id)
);

-- Индексы для ускорения поиска
CREATE INDEX idx_cart_user_id ON public.cart (user_id);
CREATE INDEX idx_cart_book_id ON public.cart (book_id);
