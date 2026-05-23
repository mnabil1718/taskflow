"use client";

import { useAppForm } from "@/lib/app-form";
import { useUpdateTask } from "@/hooks/use-tasks";
import { editTaskSchema } from "@/schemas/task.schema";
import { FormDialog } from "@/components/form-dialog";
import type { ProjectMember, Task } from "@/lib/types";

function toDateInputValue(iso: string | undefined): string {
    if (!iso) return "";
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return "";
    const tzOffsetMs = d.getTimezoneOffset() * 60_000;
    return new Date(d.getTime() - tzOffsetMs).toISOString().slice(0, 10);
}

const priorityOptions = [
    { value: "low", label: "Low" },
    { value: "medium", label: "Medium" },
    { value: "high", label: "High" },
];

const statusOptions = [
    { value: "todo", label: "To Do" },
    { value: "in_progress", label: "In Progress" },
    { value: "done", label: "Done" },
];

interface EditTaskDialogProps {
    task: Task;
    projectId: string;
    members: ProjectMember[];
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function EditTaskDialog({
    task,
    projectId,
    members,
    open,
    onOpenChange,
}: EditTaskDialogProps) {
    const updateTask = useUpdateTask(projectId);

    const assigneeOptions = [
        { value: "", label: "Unassigned" },
        ...members.map((m) => ({ value: m.user_id, label: m.name, description: m.email })),
    ];

    const form = useAppForm({
        defaultValues: {
            title: task.title,
            description: task.description ?? "",
            status: task.status as "todo" | "in_progress" | "done",
            priority: task.priority as "low" | "medium" | "high",
            assignee_id: task.assignee_id ?? "",
            due_date: toDateInputValue(task.due_date),
        },
        onSubmit: async ({ value }) => {
            try {
                await updateTask.mutateAsync({
                    id: task.id,
                    data: {
                        title: value.title.trim(),
                        description: value.description.trim() || undefined,
                        status: value.status,
                        priority: value.priority,
                        assignee_id: value.assignee_id || null,
                        due_date: value.due_date
                            ? new Date(`${value.due_date}T00:00:00Z`).toISOString()
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
        <FormDialog
            open={open}
            onOpenChange={onOpenChange}
            title="Edit task"
            description="Update the task details."
            submitLabel="Save changes"
            form={form}
        >
            <form.AppField
                name="title"
                validators={{
                    onChange: editTaskSchema.shape.title,
                    onBlur: editTaskSchema.shape.title,
                }}
                children={(field) => (
                    <field.InputField
                        label="Title"
                        placeholder="e.g. Set up CI pipeline"
                        autoComplete="off"
                        required
                    />
                )}
            />

            <form.AppField
                name="description"
                validators={{
                    onChange: editTaskSchema.shape.description,
                    onBlur: editTaskSchema.shape.description,
                }}
                children={(field) => (
                    <field.TextareaField
                        label="Description"
                        placeholder="Describe what needs to be done"
                        rows={3}
                    />
                )}
            />

            <div className="grid grid-cols-2 gap-4">
                <form.AppField
                    name="status"
                    validators={{
                        onChange: editTaskSchema.shape.status,
                        onBlur: editTaskSchema.shape.status,
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
                    name="priority"
                    validators={{
                        onChange: editTaskSchema.shape.priority,
                        onBlur: editTaskSchema.shape.priority,
                    }}
                    children={(field) => (
                        <field.SelectField
                            label="Priority"
                            options={priorityOptions}
                            required
                        />
                    )}
                />
            </div>

            <form.AppField
                name="assignee_id"
                children={(field) => (
                    <field.SelectField
                        label="Assignee"
                        options={assigneeOptions}
                        placeholder="Unassigned"
                    />
                )}
            />

            <form.AppField
                name="due_date"
                validators={{
                    onChange: editTaskSchema.shape.due_date,
                    onBlur: editTaskSchema.shape.due_date,
                }}
                children={(field) => (
                    <field.InputField
                        label="Due date"
                        type="date"
                        desc="Optional — leave blank for none"
                    />
                )}
            />
        </FormDialog>
    );
}
