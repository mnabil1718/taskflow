import { useQuery } from "@tanstack/react-query";
import { projectsApi } from "@/lib/api/projects";

export const projectKeys = {
    all: ["projects"] as const,
    detail: (id: string) => ["projects", id] as const,
    members: (id: string) => ["projects", id, "members"] as const,
};

export function useProjects() {
    return useQuery({
        queryKey: projectKeys.all,
        queryFn: projectsApi.list,
        staleTime: 60 * 1000,
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
