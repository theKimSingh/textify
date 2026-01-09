import { Worker } from "bullmq";
import { Redis } from "ioredis";
import { db } from "./db";
import { ocrResults } from "./schema";
import { eq } from "drizzle-orm";

const connection = new Redis({ maxRetriesPerRequest: null });
const worker = new Worker("ocr-tasks", async (job) => {
  const { imageUrl } = job.data;

  try {
    const response = await fetch("http://localhost:8866/predict/ocr_system", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ images: [imageUrl] })
    });
    const result = await response.json();
    const textContent = result.results[0].map((res: any) => res.text).join(" ");
    await db.update(ocrResults)
      .set({ 
        extractedText: textContent, 
        status: "completed" 
      })
      .where(eq(ocrResults.jobId, job.id!));

  } catch (error) {
    await db.update(ocrResults)
      .set({ status: "failed" })
      .where(eq(ocrResults.jobId, job.id!));
      
    throw error;
  }
}, { connection });