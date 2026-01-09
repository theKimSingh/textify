import { Elysia, t } from "elysia";
import { Queue } from "bullmq";
import { Redis } from "ioredis";
import { db } from "./db";
import { ocrResults } from "./schema";
import { eq } from "drizzle-orm";

const redisConnection = new Redis(process.env.REDIS_URL || "redis://localhost:6379", {
  maxRetriesPerRequest: null,
});

const ocrQueue = new Queue("ocr-tasks", { connection: redisConnection });
const app = new Elysia()
  .get("/", () => ({ status: "online", service: "ocr-api" }))

  /**
   * POST /textify
   * Accepts an image URL, creates a DB record, and queues the OCR task.
   */
  .post("/textify", async ({ body, set }) => {
    const { imageUrl } = body;

    try {
      const job = await ocrQueue.add("process-image", { imageUrl });
      await db.insert(ocrResults).values({
        jobId: job.id!,
        imageUrl: imageUrl,
        status: "pending",
      });

      set.status = 201;
      return {
        success: true,
        jobId: job.id,
        message: "Image queued for OCR processing",
      };
    } catch (error) {
      set.status = 500;
      return { success: false, error: "Failed to queue task" };
    }
  }, {
    body: t.Object({
      imageUrl: t.String({ format: 'uri' }),
    })
  })

  /**
   * GET /results/:jobId
   * Retrieves the current status or extracted text from Postgres.
   */
  .get("/results/:jobId", async ({ params, set }) => {
    const [record] = await db
      .select()
      .from(ocrResults)
      .where(eq(ocrResults.jobId, params.jobId))
      .limit(1);

    if (!record) {
      set.status = 404;
      return { error: "Job ID not found" };
    }

    return record;
  })

  .listen(3000);

console.log(
  `🦊 Elysia API is running at ${app.server?.hostname}:${app.server?.port}`
);