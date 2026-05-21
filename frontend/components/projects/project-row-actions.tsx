"use client";

import { useState } from "react";
import { Archive, ArchiveRestore, MoreHorizontal, Pencil, Trash2 } from "lucide-react";

import { useAuth } from "@/lib/auth-context";
import { useDeleteProject, useUpdateProject } from "@/hooks/use-projects";
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
import { EditProjectDialog } from "@/components/projects/edit-project-dialog";
import type { Project } from "@/lib/types";

type ActiveDialog = "edit" | "delete" | null;

interface ProjectRowActionsProps {
    project: Project;
}

export function ProjectRowActions({ project }: ProjectRowActionsProps) {
    const { user } = useAuth();
    const [activeDialog, setActiveDialog] = useState<ActiveDialog>(null);
    const deleteProject = useDeleteProject();
    const updateProject = useUpdateProject();

    const isOwner = user?.id === project.owner_id;
    if (!isOwner) return null;

    const handleDelete = async () => {
        await deleteProject.mutateAsync(project.id);
        setActiveDialog(null);
    };

    const isArchived = project.status === "archived";
    const handleToggleStatus = () => {
        updateProject.mutate({
            id: project.id,
            data: {
                name: project.name,
                description: project.description,
                deadline: project.deadline,
                status: isArchived ? "active" : "archived",
            },
        });
    };

    return (
        <>
            <DropdownMenu>
                <DropdownMenuTrigger
                    render={
                        <Button variant="ghost" size="icon-sm" aria-label={`Actions for ${project.name}`}>
                            <MoreHorizontal className="size-4" />
                        </Button>
                    }
                />
                <DropdownMenuContent align="end" className="min-w-36">
                    <DropdownMenuItem onClick={() => setActiveDialog("edit")}>
                        <Pencil />
                        Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                        onClick={handleToggleStatus}
                        disabled={updateProject.isPending}
                    >
                        {isArchived ? <ArchiveRestore /> : <Archive />}
                        {isArchived ? "Restore" : "Archive"}
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
                <EditProjectDialog
                    project={project}
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
                            Delete &ldquo;{project.name}&rdquo;?
                        </AlertDialogTitle>
                        <AlertDialogDescription>
                            This will remove the project from your list. It can be
                            restored by an administrator.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={deleteProject.isPending}>
                            Cancel
                        </AlertDialogCancel>
                        <AlertDialogAction
                            onClick={(e) => {
                                e.preventDefault();
                                handleDelete();
                            }}
                            disabled={deleteProject.isPending}
                            className="bg-destructive text-white hover:bg-destructive/90"
                        >
                            {deleteProject.isPending ? "Deleting…" : "Delete"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </>
    );
}
