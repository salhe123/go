alter table "public"."events" rename column "featured_image_url" to "featured_image_index";
ALTER TABLE "public"."events" ALTER COLUMN "featured_image_index" TYPE integer;
