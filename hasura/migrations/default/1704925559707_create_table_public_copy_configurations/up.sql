CREATE TABLE "public"."copy_configurations" ("project_id" uuid NOT NULL, "brand_tone" text, "perspective" text DEFAULT '1st', "character_count" integer DEFAULT 150, "template_type" text DEFAULT 'list', "tone" text DEFAULT 'humorous', "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("project_id") , FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON UPDATE cascade ON DELETE cascade);COMMENT ON TABLE "public"."copy_configurations" IS E'Additional fields necessary to configure the copy generator';
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
CREATE TRIGGER "set_public_copy_configurations_updated_at"
BEFORE UPDATE ON "public"."copy_configurations"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_copy_configurations_updated_at" ON "public"."copy_configurations"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
