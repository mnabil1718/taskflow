"use client";

import { Loader2 } from "lucide-react";

import { useAppForm } from "@/lib/app-form";
import {
    useAddMember,
    useProjectMembers,
    useRemoveMember,
    useUpdateProject,
} from "@/hooks/use-projects";
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
import {
    MemberInvitePicker,
    type InvitedUser,
} from "@/components/projects/member-invite-picker";
import type { Project, ProjectMember } from "@/lib/types";

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

function toInvitedUsers(members: ProjectMember[]): InvitedUser[] {
    return members
        .filter((m) => m.role !== "owner")
        .map((m) => ({
            user_id: m.user_id,
            role: m.role,
            name: m.name,
            email: m.email,
        }));
}

interface EditProjectDialogProps {
    project: Project;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function EditProjectDialog({ project, open, onOpenChange }: EditProjectDialogProps) {
    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle>Edit project</DialogTitle>
                    <DialogDescription>
                        Update the project details and member list.
                    </DialogDescription>
                </DialogHeader>
                <EditProjectFormBody project={project} onClose={() => onOpenChange(false)} />
            </DialogContent>
        </Dialog>
    );
}

interface EditProjectFormBodyProps {
    project: Project;
    onClose: () => void;
}

// Renders the actual form only once we've fetched the existing member list,
// so the picker initialises with chips for every current member instead of
// briefly showing empty and then re-syncing.
function EditProjectFormBody({ project, onClose }: EditProjectFormBodyProps) {
    const { data: members, isLoading } = useProjectMembers(project.id);

    if (isLoading || !members) {
        return (
            <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
                <Loader2 className="mr-2 size-4 animate-spin" />
                Loading members…
            </div>
        );
    }

    return <EditProjectForm project={project} initialMembers={toInvitedUsers(members)} onClose={onClose} />;
}

interface EditProjectFormProps {
    project: Project;
    initialMembers: InvitedUser[];
    onClose: () => void;
}

function EditProjectForm({ project, initialMembers, onClose }: EditProjectFormProps) {
    const updateProject = useUpdateProject();
    const addMember = useAddMember(project.id);
    const removeMember = useRemoveMember(project.id);

    const form = useAppForm({
        defaultValues: {
            name: project.name,
            description: project.description ?? "",
            deadline: toDateInputValue(project.deadline),
            status: project.status as "active" | "archived",
            members: initialMembers,
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

                const oldIds = new Set(initialMembers.map((m) => m.user_id));
                const newIds = new Set(value.members.map((m) => m.user_id));
                const toAdd = value.members.filter((m) => !oldIds.has(m.user_id));
                const toRemove = initialMembers.filter((m) => !newIds.has(m.user_id));

                // Run member deltas in parallel; allSettled so a single failure
                // (e.g. "already a member" race) doesn't abort the rest.
                await Promise.allSettled([
                    ...toAdd.map((m) =>
                        addMember.mutateAsync({ user_id: m.user_id, role: m.role })
                    ),
                    ...toRemove.map((m) => removeMember.mutateAsync(m.user_id)),
                ]);

                onClose();
            } catch {
                // toasted by api interceptor
            }
        },
    });

    return (
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

                <form.AppField
                    name="members"
                    children={(field) => (
                        <MemberInvitePicker
                            value={field.state.value}
                            onChange={field.handleChange}
                            label="Members"
                            desc="Add or remove project members"
                        />
                    )}
                />

                <DialogFooter>
                    <Button
                        type="button"
                        variant="outline"
                        onClick={onClose}
                        disabled={updateProject.isPending}
                    >
                        Cancel
                    </Button>
                    <form.SubmitButton className="sm:w-auto">Save changes</form.SubmitButton>
                </DialogFooter>
            </form>
        </form.AppForm>
    );
}
