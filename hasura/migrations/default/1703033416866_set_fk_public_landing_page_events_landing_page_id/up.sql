alter table "public"."landing_page_events" drop constraint "landing_page_events_landing_page_id_fkey",
  add constraint "landing_page_events_landing_page_id_fkey"
  foreign key ("landing_page_id")
  references "public"."landing_pages"
  ("id") on update cascade on delete cascade;
