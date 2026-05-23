"use client";

import { useState } from "react";
import { CheckCircle2, Circle, CircleDashed, MoreHorizontal, Pencil, Trash2 } from "lucide-react";

import { useDeleteTask, useUpdateTaskStatus } from "@/hooks/use-tasks";
import { useProjectMembers } from "@/hooks/use-projects";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { EditTaskDialog } from "@/components/tasks/edit-task-dialog";
import type { ProjectMember, Task, TaskStatus } from "@/lib/types";

type ActiveDialog = "edit" | "delete" | null;

interface TaskRowActionsProps {
    task: Task;
    projectId: string;
    /** Pass pre-loaded members to avoid an extra fetch; omit to auto-fetch. */
    members?: ProjectMember[];
}

// Three quick status items rendered at the top of every dropdown. Hiding
// the option that matches the current status keeps the menu compact and
// avoids the no-op confirmation toast the backend would otherwise return.
const statusItems: { value: TaskStatus; label: string; icon: React.ElementType }[] = [
    { value: "todo", label: "To Do", icon: CircleDashed },
    { value: "in_progress", label: "In Progress", icon: Circle },
    { value: "done", label: "Done", icon: CheckCircle2 },
];

export function TaskRowActions({ task, projectId, members: propMembers }: TaskRowActionsProps) {
    const [activeDialog, setActiveDialog] = useState<ActiveDialog>(null);
    const deleteTask = useDeleteTask(projectId);
    const updateStatus = useUpdateTaskStatus(projectId);
    // The edit dialog needs the project's member list for the assignee
    // picker. If a caller (e.g. the project page) already has the
    // members loaded, it passes them in to skip the extra fetch.
    const { data: fetchedMembers = [] } = useProjectMembers(
        propMembers === undefined ? projectId : ""
    );
    const members = propMembers ?? fetchedMembers;

    const handleDelete = async () => {
        await deleteTask.mutateAsync(task.id);
        setActiveDialog(null);
    };

    const handleSetStatus = async (status: TaskStatus) => {
        if (status === task.status) return;
        await updateStatus.mutateAsync({ id: task.id, status });
    };

    return (
        <>
            <DropdownMenu>
                <DropdownMenuTrigger
                    render={
                        <Button variant="ghost" size="icon-sm" aria-label={`Actions for ${task.title}`}>
                            <MoreHorizontal className="size-4" />
                        </Button>
                    }
                />
                <DropdownMenuContent align="end" className="min-w-40">
                    {statusItems
                        .filter((item) => item.value !== task.status)
                        .map((item) => (
                            <DropdownMenuItem
                                key={item.value}
                                onClick={() => handleSetStatus(item.value)}
                                disabled={updateStatus.isPending}
                            >
                                <item.icon />
                                Set {item.label}
                            </DropdownMenuItem>
                        ))}

                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={() => setActiveDialog("edit")}>
                        <Pencil />
                        Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                        variant="destructive"
                        onClick={() => setActiveDialog("delete")}
                    >
                        <Trash2 />
                        Delete
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            {activeDialog === "edit" && (
                <EditTaskDialog
                    task={task}
                    projectId={projectId}
                    members={members}
                    open
                    onOpenChange={(open) => !open && setActiveDialog(null)}
                />
            )}

            <AlertDialog
                open={activeDialog === "delete"}
                onOpenChange={(open) => !open && setActiveDialog(null)}
            >
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>
                            Delete &ldquo;{task.title}&rdquo;?
                        </AlertDialogTitle>
                        <AlertDialogDescription>
                            This action cannot be undone.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={deleteTask.isPending}>
                            Cancel
                        </AlertDialogCancel>
                        <AlertDialogAction
                            onClick={(e) => {
                                e.preventDefault();
                                handleDelete();
                            }}
                            disabled={deleteTask.isPending}
                            className="bg-destructive text-white hover:bg-destructive/90"
                        >
                            {deleteTask.isPending ? "Deleting…" : "Delete"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </>
    );
}
