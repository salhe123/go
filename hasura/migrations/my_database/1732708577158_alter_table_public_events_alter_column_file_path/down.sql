alter table "public"."events" rename column "images" to "file_path";
ALTER TABLE "public"."events" ALTER COLUMN "file_path" TYPE ARRAY;
