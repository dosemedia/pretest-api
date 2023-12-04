ALTER TABLE "public"."facebook_creatives" ALTER COLUMN "template_id" TYPE text;
alter table "public"."facebook_creatives" rename column "template_id" to "template_name";
