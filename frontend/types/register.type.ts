import type { z } from "zod";
import type { registerSchema } from "@/schemas/register.schema";

export type RegisterFormValues = z.infer<typeof registerSchema>;
