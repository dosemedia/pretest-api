alter table "public"."landing_pages" rename column "template_name" to "template_id";
ALTER TABLE "public"."landing_pages" ALTER COLUMN "template_id" TYPE uuid;
