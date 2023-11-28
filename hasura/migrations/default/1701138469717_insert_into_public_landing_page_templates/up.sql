INSERT INTO "public"."landing_page_templates"("id", "enabled", "name", "description", "json_schema", "ui_schema", "component", "created_at", "updated_at") VALUES (E'1254928c-0f46-4565-9681-785e69d916c3', true, E'Landing Page Demo', E'For development only', '{"type":"object","required":["ctaTitle","ctaSubtitle","ctaImageUrl","ctaColor1","ctaColor2","ctaButtonBackgroundColor"],"properties":{"ctaTitle":{"type":"string","title":"CTA Title","maxLength":100},"ctaColor1":{"type":"string","title":"CTA Color 1"},"ctaColor2":{"type":"string","title":"CTA Color 2"},"ctaImageUrl":{"type":"string","title":"CTA Image Url"},"ctaSubtitle":{"type":"string","title":"CTA Subtitle"},"pollQuestions":{"type":"array","items":{"type":"object","properties":{"title":{"type":"string","title":"Question Title"},"choices":{"type":"array","items":{"type":"string"},"title":"Choice"}}},"title":"Poll Question"},"ctaButtonLabel":{"type":"string","title":"CTA Button Label","default":"Sign Up"},"ctaButtonBackgroundColor":{"type":"string","title":"CTA Button Background Color"}}}', '{"ctaTitle":{"ui:autofocus":true,"ui:classNames":"mt-5","ui:placeholder":"Title Text"},"ui:order":["ctaTitle","ctaColor1","ctaColor2","ctaImageUrl","ctaSubtitle","ctaButtonLabel","ctaButtonBackgroundColor","pollQuestions"],"ctaColor1":{"ui:field":"colorPicker","ui:classNames":"mt-5"},"ctaColor2":{"ui:field":"colorPicker","ui:classNames":"mt-5"},"ctaImageUrl":{"ui:field":"fileUrl","ui:classNames":"mt-5"},"ctaSubtitle":{"ui:classNames":"mt-5","ui:placeholder":"Subtitle Text"},"pollQuestions":{"items":{"title":{"ui:placeholder":"Question Text"},"choices":{"items":{"ui:placeholder":"Choice Text"}}}},"ctaButtonLabel":{"ui:classNames":"mt-5","ui:placeholder":"Button Label"},"ctaButtonBackgroundColor":{"ui:field":"colorPicker","ui:classNames":"mt-5"}}', E'LandingPageDemo', E'2023-11-28T02:27:49.211022+00:00', E'2023-11-28T02:27:49.211022+00:00');