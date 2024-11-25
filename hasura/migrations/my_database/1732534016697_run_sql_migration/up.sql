CREATE TABLE public.imagestore (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL,
    url TEXT NOT NULL
);
