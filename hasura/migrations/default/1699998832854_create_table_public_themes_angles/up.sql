CREATE TABLE "public"."themes_angles" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "theme_id" uuid NOT NULL, "name" text NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id") , FOREIGN KEY ("theme_id") REFERENCES "public"."projects_themes"("id") ON UPDATE cascade ON DELETE cascade);COMMENT ON TABLE "public"."themes_angles" IS E'Angles associated with a given theme';
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
CREATE TRIGGER "set_public_themes_angles_updated_at"
BEFORE UPDATE ON "public"."themes_angles"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_themes_angles_updated_at" ON "public"."themes_angles"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
CREATE EXTENSION IF NOT EXISTS pgcrypto;
