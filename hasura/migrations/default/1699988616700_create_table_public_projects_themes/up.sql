CREATE TABLE "public"."projects_themes" ("id" uuid NOT NULL, "project_id" uuid NOT NULL, "name" text NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id") , FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON UPDATE cascade ON DELETE cascade);COMMENT ON TABLE "public"."projects_themes" IS E'Themes associated with a project';
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
CREATE TRIGGER "set_public_projects_themes_updated_at"
BEFORE UPDATE ON "public"."projects_themes"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_projects_themes_updated_at" ON "public"."projects_themes"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
