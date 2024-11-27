ALTER TABLE "public"."events" ALTER COLUMN "file_path" TYPE text[];
alter table "public"."events" rename column "file_path" to "images";
