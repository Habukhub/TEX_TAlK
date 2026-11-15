-- Add slide columns to submissions table
ALTER TABLE "submissions" ADD COLUMN IF NOT EXISTS "slide_url" TEXT;
ALTER TABLE "submissions" ADD COLUMN IF NOT EXISTS "slide_file_name" VARCHAR(255);
