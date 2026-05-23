import { z } from "zod";

export const createTaskSchema = z.object({
    title: z
        .string()
        .min(1, "Title is required")
        .max(255, "Title must be at most 255 characters"),
    description: z
        .string()
        .max(5000, "Description must be at most 5000 characters"),
    priority: z.enum(["low", "medium", "high"], { message: "Pick a priority" }),
    assignee_id: z.string(),
    due_date: z
        .string()
        .refine(
            (v) => v === "" || !Number.isNaN(new Date(v).getTime()),
            "Enter a valid date"
        ),
});

export const editTaskSchema = createTaskSchema.extend({
    status: z.enum(["todo", "in_progress", "done"], { message: "Pick a status" }),
});

export const createTaskGlobalSchema = createTaskSchema.extend({
    project_id: z.string().min(1, "Project is required"),
});
