ALTER TABLE "public"."events" ALTER COLUMN "image" TYPE text[];
alter table "public"."events" rename column "image" to "file_path";
