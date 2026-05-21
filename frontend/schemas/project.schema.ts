import { z } from "zod";

// All fields hold a string in the form state — empty string means "not set"
// for the optional ones. Trimming and date conversion happen at submit time.
export const createProjectSchema = z.object({
  name: z
    .string()
    .min(1, "Name is required")
    .max(255, "Name must be at most 255 characters"),
  description: z
    .string()
    .max(2000, "Description must be at most 2000 characters"),
  deadline: z
    .string()
    .refine(
      (v) => v === "" || !Number.isNaN(new Date(v).getTime()),
      "Enter a valid date"
    ),
});

export const editProjectSchema = createProjectSchema.extend({
  status: z.enum(["active", "archived"], {
    message: "Pick a status",
  }),
});
