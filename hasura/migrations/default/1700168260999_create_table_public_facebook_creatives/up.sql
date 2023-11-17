CREATE TABLE "public"."facebook_creatives" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "project_id" uuid NOT NULL, "template_id" uuid NOT NULL, "data" jsonb NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id") , FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON UPDATE cascade ON DELETE cascade, FOREIGN KEY ("template_id") REFERENCES "public"."facebook_creative_templates"("id") ON UPDATE cascade ON DELETE restrict);
CREATE OR REPLACE FUNCTION "public"."set_current_timestamp_updated_at"()
RETURNS TRIGGER AS $$
DECLARE
  _new record;
BEGIN
  _new := NEW;
  _new."updated_at" = NOW();
  RETURN _new;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER "set_public_facebook_creatives_updated_at"
BEFORE UPDATE ON "public"."facebook_creatives"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_facebook_creatives_updated_at" ON "public"."facebook_creatives"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
CREATE EXTENSION IF NOT EXISTS pgcrypto;
