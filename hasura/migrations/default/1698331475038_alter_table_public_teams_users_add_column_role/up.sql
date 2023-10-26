alter table "public"."teams_users" add column "role" text
 not null default 'member';
