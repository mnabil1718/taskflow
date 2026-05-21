"use client";

import { useState } from "react";
import { Plus } from "lucide-react";

import { useAppForm } from "@/lib/app-form";
import { useCreateProject } from "@/hooks/use-projects";
import { createProjectSchema } from "@/schemas/project.schema";
import { Button } from "@/components/ui/button";
import { FormDialog } from "@/components/form-dialog";
import {
    MemberInvitePicker,
    type InvitedUser,
} from "@/components/projects/member-invite-picker";

// Today as YYYY-MM-DD in the user's local timezone — used as the min for the
// deadline picker so we don't allow scheduling in the past.
function todayLocalISODate(): string {
    const d = new Date();
    const tzOffsetMs = d.getTimezoneOffset() * 60_000;
    return new Date(d.getTime() - tzOffsetMs).toISOString().slice(0, 10);
}

export function CreateProjectDialog() {
    const [open, setOpen] = useState(false);
    const createProject = useCreateProject();

    const form = useAppForm({
        defaultValues: {
            name: "",
            description: "",
            deadline: "",
            members: [] as InvitedUser[],
        },
        onSubmit: async ({ value, formApi }) => {
            try {
                await createProject.mutateAsync({
                    name: value.name.trim(),
                    description: value.description.trim() || undefined,
                    // <input type="date"> emits YYYY-MM-DD; backend wants RFC3339.
                    deadline: value.deadline
                        ? new Date(`${value.deadline}T00:00:00Z`).toISOString()
                        : undefined,
                    members: value.members.length
                        ? value.members.map((m) => ({ user_id: m.user_id, role: m.role }))
                        : undefined,
                });
                formApi.reset();
                setOpen(false);
            } catch {
                // toasted by api interceptor
            }
        },
    });

    return (
        <FormDialog
            open={open}
            onOpenChange={(next) => {
                setOpen(next);
                if (!next) form.reset();
            }}
            title="Create project"
            description="Give your project a name. You can update details later."
            submitLabel="Create"
            form={form}
            trigger={
                <Button size="lg" className="px-4!">
                    <Plus className="size-4" />
                    New Project
                </Button>
            }
        >
            <form.AppField
                name="name"
                validators={{
                    onChange: createProjectSchema.shape.name,
                    onBlur: createProjectSchema.shape.name,
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
                    onChange: createProjectSchema.shape.description,
                    onBlur: createProjectSchema.shape.description,
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
                name="deadline"
                validators={{
                    onChange: createProjectSchema.shape.deadline,
                    onBlur: createProjectSchema.shape.deadline,
                }}
                children={(field) => (
                    <field.InputField
                        label="Deadline"
                        type="date"
                        min={todayLocalISODate()}
                        desc="Optional — leave blank for none"
                    />
                )}
            />

            <form.AppField
                name="members"
                children={(field) => (
                    <MemberInvitePicker
                        value={field.state.value}
                        onChange={field.handleChange}
                    />
                )}
            />
        </FormDialog>
    );
}
