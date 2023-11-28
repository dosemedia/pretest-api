alter table "public"."teams_users"
  add constraint "teams_users_role_fkey"
  foreign key ("role")
  references "public"."teams_roles"
  ("role") on update no action on delete no action;
