alter table "public"."events" rename column "file_path" to "image";
ALTER TABLE "public"."events" ALTER COLUMN "image" TYPE ARRAY;
