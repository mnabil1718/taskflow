import type { z } from "zod";
import type { loginSchema } from "@/schemas/login.schema";

export type LoginFormValues = z.infer<typeof loginSchema>;
