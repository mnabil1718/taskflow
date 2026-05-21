"use client";

import { useEffect, useRef, useState } from "react";
import { Loader2, Search, X } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useUserSearch } from "@/hooks/use-users";
import { cn } from "@/lib/utils";
import type { ProjectMemberInvite, User } from "@/lib/types";

interface InvitedUser extends ProjectMemberInvite {
    name: string;
    email: string;
}

interface MemberInvitePickerProps {
    value: InvitedUser[];
    onChange: (next: InvitedUser[]) => void;
    label?: string;
    desc?: string;
}

// Debounces a string by `delay` ms — keeps user-search traffic to one request
// per pause rather than one per keystroke.
function useDebouncedValue(value: string, delay: number): string {
    const [debounced, setDebounced] = useState(value);
    useEffect(() => {
        const id = setTimeout(() => setDebounced(value), delay);
        return () => clearTimeout(id);
    }, [value, delay]);
    return debounced;
}

export function MemberInvitePicker({
    value,
    onChange,
    label = "Invite members",
    desc = "Optional — search teammates by name or email",
}: MemberInvitePickerProps) {
    const [query, setQuery] = useState("");
    const [open, setOpen] = useState(false);
    const inputRef = useRef<HTMLInputElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const debouncedQuery = useDebouncedValue(query, 250);

    const { data, isFetching } = useUserSearch(debouncedQuery, open);

    // Close the dropdown when the user clicks outside the picker.
    useEffect(() => {
        if (!open) return;
        const onPointer = (e: MouseEvent) => {
            if (!containerRef.current?.contains(e.target as Node)) {
                setOpen(false);
            }
        };
        document.addEventListener("mousedown", onPointer);
        return () => document.removeEventListener("mousedown", onPointer);
    }, [open]);

    const selectedIds = new Set(value.map((v) => v.user_id));
    const results = (data ?? []).filter((u) => !selectedIds.has(u.id));

    const addUser = (user: User) => {
        onChange([
            ...value,
            { user_id: user.id, role: "member", name: user.name, email: user.email },
        ]);
        setQuery("");
        inputRef.current?.focus();
    };

    const removeUser = (id: string) => {
        onChange(value.filter((v) => v.user_id !== id));
    };

    const showDropdown = open && debouncedQuery.length > 0;

    return (
        <div className="space-y-1.5" ref={containerRef}>
            <Label htmlFor="member-invite-input">{label}</Label>

            {value.length > 0 && (
                <div className="flex flex-wrap gap-2 mt-3 mb-2">
                    {value.map((u) => (
                        <Badge
                            key={u.user_id}
                            variant="secondary"
                            className="gap-1 p-3!"
                        >
                            <span className="text-xs">{u.name}</span>
                            <button
                                type="button"
                                onClick={() => removeUser(u.user_id)}
                                aria-label={`Remove ${u.name}`}
                                className="inline-flex size-4 items-center justify-center rounded-sm hover:bg-foreground/10"
                            >
                                <X className="size-3" />
                            </button>
                        </Badge>
                    ))}
                </div>
            )}

            <div className="relative">
                <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                    id="member-invite-input"
                    ref={inputRef}
                    type="text"
                    value={query}
                    placeholder="Search teammates…"
                    autoComplete="off"
                    onChange={(e) => {
                        setQuery(e.target.value);
                        setOpen(true);
                    }}
                    onFocus={() => setOpen(true)}
                    className="pl-8! py-5!"
                />
                {isFetching && debouncedQuery.length > 0 && (
                    <Loader2 className="absolute right-2.5 top-1/2 size-4 -translate-y-1/2 animate-spin text-muted-foreground" />
                )}

                {showDropdown && (
                    <div
                        className={cn(
                            "absolute z-50 mt-1 w-full overflow-hidden rounded-lg bg-popover text-popover-foreground shadow-md ring-1 ring-foreground/10"
                        )}
                    >
                        {isFetching && results.length === 0 ? (
                            <div className="px-3 py-2 text-xs text-muted-foreground">
                                Searching…
                            </div>
                        ) : results.length === 0 ? (
                            <div className="px-3 py-2 text-xs text-muted-foreground">
                                No users match &ldquo;{debouncedQuery}&rdquo;
                            </div>
                        ) : (
                            <ul className="max-h-56 overflow-y-auto py-1">
                                {results.map((user) => (
                                    <li key={user.id}>
                                        <button
                                            type="button"
                                            onClick={() => addUser(user)}
                                            className="flex w-full flex-col items-start px-3 py-1.5 text-left text-sm hover:bg-accent hover:text-accent-foreground"
                                        >
                                            <span className="font-medium">{user.name}</span>
                                            <span className="text-xs text-muted-foreground">
                                                {user.email}
                                            </span>
                                        </button>
                                    </li>
                                ))}
                            </ul>
                        )}
                    </div>
                )}
            </div>

            {desc && (
                <p className="text-[0.8rem] text-muted-foreground">{desc}</p>
            )}
        </div>
    );
}

export type { InvitedUser };
