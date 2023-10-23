CREATE TABLE "public"."invitations" ("team_id" uuid NOT NULL, "email" text NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("team_id","email") , FOREIGN KEY ("team_id") REFERENCES "public"."teams"("id") ON UPDATE cascade ON DELETE cascade, CONSTRAINT "email_format" CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'::text));COMMENT ON TABLE "public"."invitations" IS E'Invitations from teams to users';