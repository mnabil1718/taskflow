"use client";

import { useRef, useState } from "react";
import { Loader2, Search, UserMinus } from "lucide-react";

import { useAddMember, useProjectMembers, useRemoveMember } from "@/hooks/use-projects";
import { useUserSearch } from "@/hooks/use-users";
import { useDebounced } from "@/hooks/use-debounced";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import type { User } from "@/lib/types";

const roleBadgeVariant: Record<string, "default" | "secondary" | "outline"> = {
    owner: "default",
    admin: "secondary",
    member: "outline",
};

interface ProjectMembersPanelProps {
    projectId: string;
    currentUserId: string;
}

export function ProjectMembersPanel({ projectId, currentUserId }: ProjectMembersPanelProps) {
    const [query, setQuery] = useState("");
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);

    const { data: members = [], isLoading } = useProjectMembers(projectId);
    const addMember = useAddMember(projectId);
    const removeMember = useRemoveMember(projectId);

    const debouncedQuery = useDebounced(query, 250);
    const { data: searchResults = [], isFetching } = useUserSearch(debouncedQuery, dropdownOpen);

    const existingIds = new Set(members.map((m) => m.user_id));
    const filteredResults = searchResults.filter((u) => !existingIds.has(u.id));

    const handleAdd = async (user: User) => {
        await addMember.mutateAsync({ user_id: user.id, role: "member" });
        setQuery("");
        setDropdownOpen(false);
    };

    const handleRemove = async (userId: string) => {
        await removeMember.mutateAsync(userId);
    };

    const showDropdown = dropdownOpen && debouncedQuery.length > 0;

    return (
        <div className="space-y-6 max-w-lg">
            {/* Add member search */}
            <div className="space-y-1.5">
                <p className="text-sm font-medium">Add member</p>
                <div className="relative" ref={containerRef}>
                    <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                        type="text"
                        value={query}
                        placeholder="Search by name or email…"
                        autoComplete="off"
                        onChange={(e) => {
                            setQuery(e.target.value);
                            setDropdownOpen(true);
                        }}
                        onFocus={() => setDropdownOpen(true)}
                        onBlur={() => setTimeout(() => setDropdownOpen(false), 150)}
                        className="pl-8! py-5!"
                    />
                    {isFetching && debouncedQuery.length > 0 && (
                        <Loader2 className="absolute right-2.5 top-1/2 size-4 -translate-y-1/2 animate-spin text-muted-foreground" />
                    )}
                    {showDropdown && (
                        <div className="absolute z-50 mt-1 w-full overflow-hidden rounded-lg bg-popover text-popover-foreground shadow-md ring-1 ring-foreground/10">
                            {isFetching && filteredResults.length === 0 ? (
                                <div className="px-3 py-2 text-xs text-muted-foreground">Searching…</div>
                            ) : filteredResults.length === 0 ? (
                                <div className="px-3 py-2 text-xs text-muted-foreground">
                                    {debouncedQuery.length > 0
                                        ? `No users match "${debouncedQuery}"`
                                        : "Type to search"}
                                </div>
                            ) : (
                                <ul className="max-h-56 overflow-y-auto py-1">
                                    {filteredResults.map((user) => (
                                        <li key={user.id}>
                                            <button
                                                type="button"
                                                onMouseDown={(e) => e.preventDefault()}
                                                onClick={() => handleAdd(user)}
                                                disabled={addMember.isPending}
                                                className={cn(
                                                    "flex w-full flex-col items-start px-3 py-1.5 text-left text-sm",
                                                    "hover:bg-accent hover:text-accent-foreground",
                                                    "disabled:opacity-50 disabled:pointer-events-none"
                                                )}
                                            >
                                                <span className="font-medium">{user.name}</span>
                                                <span className="text-xs text-muted-foreground">{user.email}</span>
                                            </button>
                                        </li>
                                    ))}
                                </ul>
                            )}
                        </div>
                    )}
                </div>
            </div>

            <Separator />

            {/* Members list */}
            <div className="space-y-1.5">
                <p className="text-sm font-medium">
                    Members
                    {members.length > 0 && (
                        <span className="ml-1.5 text-muted-foreground font-normal">({members.length})</span>
                    )}
                </p>
                {isLoading ? (
                    <div className="flex items-center gap-2 py-4 text-sm text-muted-foreground">
                        <Loader2 className="size-4 animate-spin" />
                        Loading members…
                    </div>
                ) : members.length === 0 ? (
                    <p className="text-sm text-muted-foreground py-4">No members yet.</p>
                ) : (
                    <ul className="divide-y divide-border rounded-lg border">
                        {members.map((member) => (
                            <li key={member.user_id} className="flex items-center gap-3 px-4 py-3">
                                <div className="flex-1 min-w-0">
                                    <p className="text-sm font-medium truncate">{member.name}</p>
                                    <p className="text-xs text-muted-foreground truncate">{member.email}</p>
                                </div>
                                <Badge variant={roleBadgeVariant[member.role] ?? "outline"} className="capitalize shrink-0">
                                    {member.role}
                                </Badge>
                                {member.role !== "owner" && member.user_id !== currentUserId && (
                                    <Button
                                        variant="ghost"
                                        size="icon-sm"
                                        aria-label={`Remove ${member.name}`}
                                        onClick={() => handleRemove(member.user_id)}
                                        disabled={removeMember.isPending}
                                    >
                                        <UserMinus className="size-4 text-destructive" />
                                    </Button>
                                )}
                            </li>
                        ))}
                    </ul>
                )}
            </div>
        </div>
    );
}
