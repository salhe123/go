alter table "public"."imagestore" drop constraint "fk_event",
  add constraint "imagestore_event_id_fkey"
  foreign key ("event_id")
  references "public"."events"
  ("id") on update no action on delete cascade;
