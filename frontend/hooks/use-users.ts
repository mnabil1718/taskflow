import { useQuery } from "@tanstack/react-query";
import { usersApi } from "@/lib/api/users";

export const userKeys = {
    search: (q: string) => ["users", "search", q] as const,
};

export function useUserSearch(query: string, enabled = true) {
    const trimmed = query.trim();
    return useQuery({
        queryKey: userKeys.search(trimmed),
        queryFn: () => usersApi.search(trimmed),
        staleTime: 30 * 1000,
        enabled: enabled && trimmed.length > 0,
    });
}
