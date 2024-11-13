alter table "public"."users" add constraint "users_last_name_key" unique (last_name);
alter table "public"."users" alter column "last_name" drop not null;
alter table "public"."users" add column "last_name" varchar;
