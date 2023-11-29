ALTER TABLE "public"."landing_pages" ALTER COLUMN "template_id" TYPE text;
alter table "public"."landing_pages" rename column "template_id" to "template_name";
