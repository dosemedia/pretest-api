alter table "public"."facebook_creatives" rename column "template_name" to "template_id";
ALTER TABLE "public"."facebook_creatives" ALTER COLUMN "template_id" TYPE uuid;
