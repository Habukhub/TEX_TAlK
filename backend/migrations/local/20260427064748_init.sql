-- Create "submissions" table
CREATE TABLE "submissions" (
  "id" bigserial NOT NULL,
  "full_name" character varying(255) NOT NULL,
  "email" character varying(255) NOT NULL,
  "topic_title" character varying(255) NOT NULL,
  "status" character varying(20) NULL DEFAULT 'PENDING',
  "approved_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "users" table
CREATE TABLE "users" (
  "id" bigserial NOT NULL,
  "email" character varying(255) NOT NULL,
  "name" character varying(255) NOT NULL,
  "google_id" character varying(255) NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_users_email" UNIQUE ("email")
);
-- Create "schedules" table
CREATE TABLE "schedules" (
  "id" bigserial NOT NULL,
  "submission_id" bigint NULL,
  "schedule_type" character varying(20) NULL DEFAULT 'SPEAKER',
  "start_time" character varying(5) NOT NULL,
  "end_time" character varying(5) NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_submissions_schedules" FOREIGN KEY ("submission_id") REFERENCES "submissions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_schedules_submission_id" to table: "schedules"
CREATE INDEX "idx_schedules_submission_id" ON "schedules" ("submission_id");
