-- Drop foreign key constraint
ALTER TABLE "schedules" DROP CONSTRAINT "fk_submissions_schedules";
-- Modify "submissions" table
ALTER TABLE "submissions" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE uuid USING gen_random_uuid(), ALTER COLUMN "id" SET DEFAULT gen_random_uuid();
-- Drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "submissions_id_seq";
-- Modify "schedules" table
ALTER TABLE "schedules" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE uuid USING gen_random_uuid(), ALTER COLUMN "id" SET DEFAULT gen_random_uuid(), ALTER COLUMN "submission_id" TYPE uuid USING NULL;
-- Drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "schedules_id_seq";
-- Re-add foreign key constraint
ALTER TABLE "schedules" ADD CONSTRAINT "fk_submissions_schedules" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id") ON DELETE SET NULL;
-- Modify "users" table
ALTER TABLE "users" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE uuid USING gen_random_uuid(), ALTER COLUMN "id" SET DEFAULT gen_random_uuid();
-- Drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "users_id_seq";
