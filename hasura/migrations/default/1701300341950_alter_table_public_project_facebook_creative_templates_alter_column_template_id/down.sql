alter table "public"."project_facebook_creative_templates" rename column "template_name" to "template_id";
ALTER TABLE "public"."project_facebook_creative_templates" ALTER COLUMN "template_id" TYPE uuid;
