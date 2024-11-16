ALTER TABLE "public"."events" ALTER COLUMN "featured_image_index" TYPE VARCHAR;
alter table "public"."events" rename column "featured_image_index" to "featured_image_url";
