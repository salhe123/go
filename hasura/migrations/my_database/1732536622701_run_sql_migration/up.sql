ALTER TABLE events
  ADD CONSTRAINT fk_events_organizer
  FOREIGN KEY (organizer_id)
  REFERENCES users(id) ON DELETE SET NULL;
