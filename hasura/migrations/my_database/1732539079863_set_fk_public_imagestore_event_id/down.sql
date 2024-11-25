alter table "public"."imagestore" drop constraint "imagestore_event_id_fkey",
  add constraint "fk_event"
  foreign key ("event_id")
  references "public"."events"
  ("id") on update no action on delete cascade;
