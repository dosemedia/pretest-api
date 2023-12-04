alter table "public"."facebook_creatives"
  add constraint "facebook_creatives_angle_id_fkey"
  foreign key ("angle_id")
  references "public"."themes_angles"
  ("id") on update cascade on delete cascade;
