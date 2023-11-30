alter table "public"."project_facebook_creative_templates"
  add constraint "project_facebook_creative_template_template_id_fkey"
  foreign key ("template_id")
  references "public"."facebook_creative_templates"
  ("id") on update cascade on delete cascade;
