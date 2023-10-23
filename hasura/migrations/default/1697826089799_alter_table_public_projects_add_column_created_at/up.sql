alter table "public"."projects" add column "created_at" timestamptz
 not null default now();
