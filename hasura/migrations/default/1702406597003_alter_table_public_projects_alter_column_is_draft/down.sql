alter table "public"."projects" rename column "status" to "is_draft";
alter table "public"."projects" alter column "is_draft" set default 'true';
ALTER TABLE "public"."projects" ALTER COLUMN "is_draft" TYPE boolean;
