alter table "public"."events" alter column "featured_image_url" drop not null;
alter table "public"."events" add column "featured_image_url" varchar;
