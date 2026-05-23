import { useMutation, useQuery, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import { toast } from "sonner";
import { projectsApi } from "@/lib/api/projects";
import type {
    AddMemberRequest,
    CreateProjectRequest,
    ProjectListParams,
    UpdateProjectRequest,
} from "@/lib/types";

export const projectKeys = {
    all: ["projects"] as const,
    list: (params: ProjectListParams) => ["projects", "list", params] as const,
    detail: (id: string) => ["projects", id] as const,
    members: (id: string) => ["projects", id, "members"] as const,
};

export function useProjects(params: ProjectListParams = {}) {
    const { page = 1, limit = 10 } = params;
    return useQuery({
        queryKey: projectKeys.list({ page, limit }),
        queryFn: () => projectsApi.list({ page, limit }),
        staleTime: 60 * 1000,
        placeholderData: keepPreviousData,
    });
}

export function useProject(id: string) {
    return useQuery({
        queryKey: projectKeys.detail(id),
        queryFn: () => projectsApi.getById(id),
        staleTime: 60 * 1000,
        enabled: !!id,
    });
}

export function useProjectMembers(id: string) {
    return useQuery({
        queryKey: projectKeys.members(id),
        queryFn: () => projectsApi.getMembers(id),
        staleTime: 60 * 1000,
        enabled: !!id,
    });
}

export function useCreateProject() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateProjectRequest) => projectsApi.create(data),
        onSuccess: (project) => {
            qc.invalidateQueries({ queryKey: projectKeys.all });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            toast.success(`Project "${project.name}" created`);
        },
    });
}

export function useUpdateProject() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: string; data: UpdateProjectRequest }) =>
            projectsApi.update(id, data),
        onSuccess: (project) => {
            qc.invalidateQueries({ queryKey: projectKeys.all });
            qc.invalidateQueries({ queryKey: projectKeys.detail(project.id) });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            toast.success(`Project "${project.name}" updated`);
        },
    });
}

export function useDeleteProject() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => projectsApi.delete(id),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: projectKeys.all });
            qc.invalidateQueries({ queryKey: ["tasks", "all"] });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            qc.invalidateQueries({ queryKey: ["trash"] });
            toast.success("Project deleted");
        },
    });
}

export function useAddMember(projectID: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (data: AddMemberRequest) => projectsApi.addMember(projectID, data),
        onSuccess: (member) => {
            qc.invalidateQueries({ queryKey: projectKeys.members(projectID) });
            toast.success(`${member.name} added to project`);
        },
    });
}

export function useRemoveMember(projectID: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (userID: string) => projectsApi.removeMember(projectID, userID),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: projectKeys.members(projectID) });
            toast.success("Member removed");
        },
    });
}

export function useBulkDeleteProjects() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (ids: string[]) => projectsApi.bulkDelete(ids),
        onSuccess: ({ deleted_count }) => {
            qc.invalidateQueries({ queryKey: projectKeys.all });
            qc.invalidateQueries({ queryKey: ["tasks", "all"] });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            qc.invalidateQueries({ queryKey: ["trash"] });
            toast.success(
                `Deleted ${deleted_count} project${deleted_count === 1 ? "" : "s"}`
            );
        },
    });
}
