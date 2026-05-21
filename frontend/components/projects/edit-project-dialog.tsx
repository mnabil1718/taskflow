"use client";

import { useAppForm } from "@/lib/app-form";
import { useUpdateProject } from "@/hooks/use-projects";
import { editProjectSchema } from "@/schemas/project.schema";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import type { Project } from "@/lib/types";

const statusOptions = [
    { value: "active", label: "Active" },
    { value: "archived", label: "Archived" },
];

// Backend deadline is a full RFC3339 timestamp; the <input type="date">
// only accepts YYYY-MM-DD, so slice the day off and convert back at submit.
function toDateInputValue(iso: string | undefined): string {
    if (!iso) return "";
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return "";
    const tzOffsetMs = d.getTimezoneOffset() * 60_000;
    return new Date(d.getTime() - tzOffsetMs).toISOString().slice(0, 10);
}

interface EditProjectDialogProps {
    project: Project;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function EditProjectDialog({ project, open, onOpenChange }: EditProjectDialogProps) {
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
                        description: value.description.trim(),
                        status: value.status,
                        deadline: value.deadline
                            ? new Date(`${value.deadline}T00:00:00Z`).toISOString()
                            : undefined,
                    },
                });
                onOpenChange(false);
            } catch {
                // toasted by api interceptor
            }
        },
    });

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle>Edit project</DialogTitle>
                    <DialogDescription>
                        Update the project details and save your changes.
                    </DialogDescription>
                </DialogHeader>

                <form.AppForm>
                    <form
                        onSubmit={(e) => {
                            e.preventDefault();
                            form.handleSubmit();
                        }}
                        className="space-y-4"
                    >
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

                        <DialogFooter>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => onOpenChange(false)}
                                disabled={updateProject.isPending}
                            >
                                Cancel
                            </Button>
                            <form.SubmitButton className="sm:w-auto">Save changes</form.SubmitButton>
                        </DialogFooter>
                    </form>
                </form.AppForm>
            </DialogContent>
        </Dialog>
    );
}
