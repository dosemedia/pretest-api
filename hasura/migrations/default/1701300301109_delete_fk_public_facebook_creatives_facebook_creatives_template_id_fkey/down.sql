alter table "public"."facebook_creatives"
  add constraint "facebook_creatives_template_id_fkey"
  foreign key ("template_id")
  references "public"."facebook_creative_templates"
  ("id") on update cascade on delete restrict;
