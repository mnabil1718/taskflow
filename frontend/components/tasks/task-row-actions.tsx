"use client";

import { useState } from "react";
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react";

import { useDeleteTask } from "@/hooks/use-tasks";
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
import type { ProjectMember, Task } from "@/lib/types";

type ActiveDialog = "edit" | "delete" | null;

interface TaskRowActionsProps {
    task: Task;
    projectId: string;
    /** Pass pre-loaded members to avoid an extra fetch; omit to auto-fetch. */
    members?: ProjectMember[];
}

export function TaskRowActions({ task, projectId, members: propMembers }: TaskRowActionsProps) {
    const [activeDialog, setActiveDialog] = useState<ActiveDialog>(null);
    const deleteTask = useDeleteTask(projectId);
    const { data: fetchedMembers = [] } = useProjectMembers(
        propMembers === undefined ? projectId : ""
    );
    const members = propMembers ?? fetchedMembers;

    const handleDelete = async () => {
        await deleteTask.mutateAsync(task.id);
        setActiveDialog(null);
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
                <DropdownMenuContent align="end" className="min-w-36">
                    <DropdownMenuItem onClick={() => setActiveDialog("edit")}>
                        <Pencil />
                        Edit
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
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
