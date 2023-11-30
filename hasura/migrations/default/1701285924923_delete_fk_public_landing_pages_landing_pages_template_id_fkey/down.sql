alter table "public"."landing_pages"
  add constraint "landing_pages_template_id_fkey"
  foreign key ("template_id")
  references "public"."landing_page_templates"
  ("id") on update cascade on delete restrict;
