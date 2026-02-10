import { pgTable, uuid, text, varchar, timestamp } from "drizzle-orm/pg-core";

export const ocrResults = pgTable("ocr_results", {
  id: uuid("id").primaryKey().defaultRandom(),
  jobId: varchar("job_id").notNull(),     // BullMQ Job ID
  imageUrl: text("image_url").notNull(),
  extractedText: text("extracted_text"),  // The final result
  status: varchar("status").default("pending"), // pending, completed, failed
  createdAt: timestamp("created_at").defaultNow(),
});