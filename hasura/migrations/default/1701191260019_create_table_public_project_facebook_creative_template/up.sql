CREATE TABLE "public"."project_facebook_creative_template" ("id" uuid NOT NULL, "project_id" uuid NOT NULL, "template_id" uuid NOT NULL, "data" jsonb NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id","template_id") , FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON UPDATE cascade ON DELETE cascade, FOREIGN KEY ("template_id") REFERENCES "public"."facebook_creative_templates"("id") ON UPDATE cascade ON DELETE cascade, UNIQUE ("id"));COMMENT ON TABLE "public"."project_facebook_creative_template" IS E'A middle layer between facebook_creative_templates and facebook_creatives that is a mutable template';
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
CREATE TRIGGER "set_public_project_facebook_creative_template_updated_at"
BEFORE UPDATE ON "public"."project_facebook_creative_template"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_project_facebook_creative_template_updated_at" ON "public"."project_facebook_creative_template"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
