ALTER TABLE public.imagestore
ADD CONSTRAINT fk_event_id FOREIGN KEY (event_id) REFERENCES public.events (id) ON DELETE CASCADE;
