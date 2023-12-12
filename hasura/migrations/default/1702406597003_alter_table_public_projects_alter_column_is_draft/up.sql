ALTER TABLE "public"."projects" ALTER COLUMN "is_draft" TYPE text;
alter table "public"."projects" alter column "is_draft" set default 'draft';
alter table "public"."projects" rename column "is_draft" to "status";
