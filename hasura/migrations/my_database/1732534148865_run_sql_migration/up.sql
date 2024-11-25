ALTER TABLE public.transactions
ADD CONSTRAINT fk_event_id FOREIGN KEY (event_id) REFERENCES public.events (id) ON DELETE CASCADE;

ALTER TABLE public.transactions
ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE CASCADE;
