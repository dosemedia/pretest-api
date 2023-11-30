ALTER TABLE "public"."project_facebook_creative_templates" ALTER COLUMN "template_id" TYPE text;
alter table "public"."project_facebook_creative_templates" rename column "template_id" to "template_name";
