BEGIN TRANSACTION;
ALTER TABLE "public"."project_facebook_creative_templates" DROP CONSTRAINT "project_facebook_creative_template_pkey";

ALTER TABLE "public"."project_facebook_creative_templates"
    ADD CONSTRAINT "project_facebook_creative_template_pkey" PRIMARY KEY ("id");
COMMIT TRANSACTION;
