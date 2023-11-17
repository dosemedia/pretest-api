CREATE TABLE "public"."facebook_creative_templates" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "enabled" boolean NOT NULL DEFAULT True, "name" text NOT NULL, "description" text, "json_schema" jsonb NOT NULL, "ui_schema" jsonb, "creatomate_template_id" text NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id") );
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
CREATE TRIGGER "set_public_facebook_creative_templates_updated_at"
BEFORE UPDATE ON "public"."facebook_creative_templates"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_facebook_creative_templates_updated_at" ON "public"."facebook_creative_templates"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
CREATE EXTENSION IF NOT EXISTS pgcrypto;
