"use client";

import { useState } from "react";
import { Plus } from "lucide-react";

import { useAppForm } from "@/lib/app-form";
import { useCreateTask } from "@/hooks/use-tasks";
import { useProjects, useProjectMembers } from "@/hooks/use-projects";
import { createTaskSchema } from "@/schemas/task.schema";
import { Button } from "@/components/ui/button";
import { FormDialog } from "@/components/form-dialog";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

function todayLocalISODate(): string {
    const d = new Date();
    const tzOffsetMs = d.getTimezoneOffset() * 60_000;
    return new Date(d.getTime() - tzOffsetMs).toISOString().slice(0, 10);
}

const priorityOptions = [
    { value: "low", label: "Low" },
    { value: "medium", label: "Medium" },
    { value: "high", label: "High" },
];

export function CreateTaskGlobalDialog() {
    const [open, setOpen] = useState(false);
    const [selectedProjectId, setSelectedProjectId] = useState("");
    const [projectError, setProjectError] = useState("");

    const { data: projectsData } = useProjects({ page: 1, limit: 100 });
    const projects = projectsData?.items ?? [];

    const { data: members = [] } = useProjectMembers(selectedProjectId);
    const createTask = useCreateTask(selectedProjectId);

    const assigneeOptions = [
        { value: "", label: "Unassigned" },
        ...members.map((m) => ({ value: m.user_id, label: m.name })),
    ];

    const form = useAppForm({
        defaultValues: {
            title: "",
            description: "",
            priority: "medium" as "low" | "medium" | "high",
            assignee_id: "",
            due_date: "",
        },
        onSubmit: async ({ value, formApi }) => {
            if (!selectedProjectId) {
                setProjectError("Project is required");
                return;
            }
            try {
                await createTask.mutateAsync({
                    title: value.title.trim(),
                    description: value.description.trim() || undefined,
                    priority: value.priority,
                    assignee_id: value.assignee_id || null,
                    due_date: value.due_date
                        ? new Date(`${value.due_date}T00:00:00Z`).toISOString()
                        : undefined,
                });
                formApi.reset();
                setSelectedProjectId("");
                setProjectError("");
                setOpen(false);
            } catch {
                // toasted by api interceptor
            }
        },
    });

    const handleClose = (next: boolean) => {
        setOpen(next);
        if (!next) {
            form.reset();
            setSelectedProjectId("");
            setProjectError("");
        }
    };

    const handleProjectChange = (pid: string | null) => {
        setSelectedProjectId(pid ?? "");
        setProjectError("");
        form.setFieldValue("assignee_id", "");
    };

    return (
        <FormDialog
            open={open}
            onOpenChange={handleClose}
            title="Create task"
            description="Add a new task to a project."
            submitLabel="Create"
            form={form}
            trigger={
                <Button size="lg" className="px-4!">
                    <Plus className="size-4" />
                    New Task
                </Button>
            }
        >
            <div className="space-y-1.5">
                <Label>
                    Project
                    <span className="ml-0.5 text-destructive">*</span>
                </Label>
                <Select
                    value={selectedProjectId || undefined}
                    onValueChange={handleProjectChange}
                >
                    <SelectTrigger aria-invalid={!!projectError}>
                        <SelectValue placeholder="Select a project" />
                    </SelectTrigger>
                    <SelectContent>
                        {projects.length === 0 ? (
                            <div className="px-3 py-2 text-xs text-muted-foreground">
                                No projects available
                            </div>
                        ) : (
                            projects.map((p) => (
                                <SelectItem key={p.id} value={p.id}>
                                    {p.name}
                                </SelectItem>
                            ))
                        )}
                    </SelectContent>
                </Select>
                {projectError && (
                    <p className="text-[0.8rem] font-medium text-destructive">
                        {projectError}
                    </p>
                )}
            </div>

            <form.AppField
                name="title"
                validators={{
                    onChange: createTaskSchema.shape.title,
                    onBlur: createTaskSchema.shape.title,
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
                    onChange: createTaskSchema.shape.description,
                    onBlur: createTaskSchema.shape.description,
                }}
                children={(field) => (
                    <field.TextareaField
                        label="Description"
                        placeholder="Describe what needs to be done"
                        rows={3}
                    />
                )}
            />

            <form.AppField
                name="priority"
                validators={{
                    onChange: createTaskSchema.shape.priority,
                    onBlur: createTaskSchema.shape.priority,
                }}
                children={(field) => (
                    <field.SelectField
                        label="Priority"
                        options={priorityOptions}
                        required
                    />
                )}
            />

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
                    onChange: createTaskSchema.shape.due_date,
                    onBlur: createTaskSchema.shape.due_date,
                }}
                children={(field) => (
                    <field.InputField
                        label="Due date"
                        type="date"
                        min={todayLocalISODate()}
                        desc="Optional — leave blank for none"
                    />
                )}
            />
        </FormDialog>
    );
}
