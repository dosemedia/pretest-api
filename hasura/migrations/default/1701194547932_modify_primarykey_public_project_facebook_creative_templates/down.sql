alter table "public"."project_facebook_creative_templates" drop constraint "project_facebook_creative_templates_pkey";
alter table "public"."project_facebook_creative_templates"
    add constraint "project_facebook_creative_template_pkey"
    primary key ("template_id", "id");
