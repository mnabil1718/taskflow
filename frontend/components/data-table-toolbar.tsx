"use client";

import { Search, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface DataTableToolbarProps {
    /** Current value of the search input (live, not debounced). */
    searchValue: string;
    onSearchChange: (value: string) => void;
    searchPlaceholder?: string;
    /** Optional filter controls rendered to the right of the search box. */
    children?: React.ReactNode;
}

export function DataTableToolbar({
    searchValue,
    onSearchChange,
    searchPlaceholder = "Search…",
    children,
}: DataTableToolbarProps) {
    return (
        <div className="flex flex-wrap items-center gap-3">
            <div className="relative flex-1 min-w-48 max-w-sm">
                <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                    value={searchValue}
                    onChange={(e) => onSearchChange(e.target.value)}
                    placeholder={searchPlaceholder}
                    className="pl-8!"
                />
                {searchValue && (
                    <Button
                        variant="ghost"
                        size="icon-sm"
                        className="absolute right-1 top-1/2 -translate-y-1/2"
                        onClick={() => onSearchChange("")}
                        aria-label="Clear search"
                    >
                        <X className="size-3.5" />
                    </Button>
                )}
            </div>
            {children}
        </div>
    );
}
