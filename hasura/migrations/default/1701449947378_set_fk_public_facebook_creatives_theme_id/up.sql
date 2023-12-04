alter table "public"."facebook_creatives"
  add constraint "facebook_creatives_theme_id_fkey"
  foreign key ("theme_id")
  references "public"."projects_themes"
  ("id") on update cascade on delete cascade;
