alter table "public"."projects" alter column "draft_step" drop not null;
alter table "public"."projects" add column "draft_step" text;
