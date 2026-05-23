"use client";

import { useAppForm } from "@/lib/app-form";
import { useUpdateProject } from "@/hooks/use-projects";
import { editProjectSchema } from "@/schemas/project.schema";
import { Button } from "@/components/ui/button";
import type { Project } from "@/lib/types";

const statusOptions = [
    { value: "active", label: "Active" },
    { value: "archived", label: "Archived" },
];

function toDateInputValue(iso: string | undefined): string {
    if (!iso) return "";
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return "";
    const tzOffsetMs = d.getTimezoneOffset() * 60_000;
    return new Date(d.getTime() - tzOffsetMs).toISOString().slice(0, 10);
}

export function ProjectMetadataForm({ project }: { project: Project }) {
    const updateProject = useUpdateProject();

    const form = useAppForm({
        defaultValues: {
            name: project.name,
            description: project.description ?? "",
            deadline: toDateInputValue(project.deadline),
            status: project.status as "active" | "archived",
        },
        onSubmit: async ({ value }) => {
            try {
                await updateProject.mutateAsync({
                    id: project.id,
                    data: {
                        name: value.name.trim(),
                        description: value.description.trim() || undefined,
                        status: value.status,
                        deadline: value.deadline
                            ? new Date(`${value.deadline}T00:00:00Z`).toISOString()
                            : undefined,
                    },
                });
            } catch {
                // toasted by api interceptor
            }
        },
    });

    return (
        <form.AppForm>
            <div className="space-y-4 max-w-lg">
                <form.AppField
                    name="name"
                    validators={{
                        onChange: editProjectSchema.shape.name,
                        onBlur: editProjectSchema.shape.name,
                    }}
                    children={(field) => (
                        <field.InputField
                            label="Name"
                            placeholder="e.g. Q4 Roadmap"
                            autoComplete="off"
                            required
                        />
                    )}
                />

                <form.AppField
                    name="description"
                    validators={{
                        onChange: editProjectSchema.shape.description,
                        onBlur: editProjectSchema.shape.description,
                    }}
                    children={(field) => (
                        <field.TextareaField
                            label="Description"
                            placeholder="What is this project about?"
                            rows={3}
                        />
                    )}
                />

                <form.AppField
                    name="status"
                    validators={{
                        onChange: editProjectSchema.shape.status,
                        onBlur: editProjectSchema.shape.status,
                    }}
                    children={(field) => (
                        <field.SelectField
                            label="Status"
                            options={statusOptions}
                            required
                        />
                    )}
                />

                <form.AppField
                    name="deadline"
                    validators={{
                        onChange: editProjectSchema.shape.deadline,
                        onBlur: editProjectSchema.shape.deadline,
                    }}
                    children={(field) => (
                        <field.InputField
                            label="Deadline"
                            type="date"
                            desc="Optional — leave blank for none"
                        />
                    )}
                />

                <form.Subscribe
                    selector={(state) => state.isSubmitting}
                    children={(isSubmitting) => (
                        <Button
                            type="button"
                            onClick={() => form.handleSubmit()}
                            disabled={isSubmitting}
                        >
                            {isSubmitting ? "Saving…" : "Save changes"}
                        </Button>
                    )}
                />
            </div>
        </form.AppForm>
    );
}
