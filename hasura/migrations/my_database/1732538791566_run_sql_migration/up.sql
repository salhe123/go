ALTER TABLE public.imagestore
ADD CONSTRAINT fk_event FOREIGN KEY (event_id)
REFERENCES public.events (id) ON DELETE CASCADE;
