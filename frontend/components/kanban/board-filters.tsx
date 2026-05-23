"use client";

import { Search, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import {
    DEFAULT_BOARD_FILTER,
    type BoardFilter,
    type SortKey,
} from "@/lib/board-filters";

// Centralised label table so the toolbar and any future readout
// (e.g. a badge that explains the active sort) read from one source.
const sortLabels: Record<SortKey, string> = {
    position: "Manual order",
    title: "Title (A → Z)",
    deadline: "Deadline (soonest)",
    priority: "Priority (highest)",
};

interface BoardFiltersProps {
    value: BoardFilter;
    onChange: (next: BoardFilter) => void;
    /** "board" renders the full-width toolbar above the columns;
     * "column" renders a compact inline version inside each column. */
    scope: "board" | "column";
    placeholder?: string;
    className?: string;
}

export function BoardFilters({
    value,
    onChange,
    scope,
    placeholder,
    className,
}: BoardFiltersProps) {
    const isDirty =
        value.search !== DEFAULT_BOARD_FILTER.search ||
        value.sortBy !== DEFAULT_BOARD_FILTER.sortBy;

    return (
        <div
            className={cn(
                "flex items-center gap-2",
                scope === "board" ? "" : "text-xs",
                className
            )}
        >
            <div className="relative flex-1 min-w-0">
                <Search
                    className={cn(
                        "pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground",
                        scope === "board" ? "size-4" : "size-3.5"
                    )}
                />
                <Input
                    type="text"
                    value={value.search}
                    placeholder={
                        placeholder ?? (scope === "board" ? "Search board…" : "Search…")
                    }
                    onChange={(e) => onChange({ ...value, search: e.target.value })}
                    autoComplete="off"
                    className={cn(
                        scope === "board" ? "pl-8 h-9" : "pl-7 h-7 text-xs"
                    )}
                />
            </div>

            <Select
                value={value.sortBy}
                onValueChange={(v) => onChange({ ...value, sortBy: (v ?? "position") as SortKey })}
            >
                <SelectTrigger
                    className={cn(
                        scope === "board" ? "w-44 h-9" : "w-36 h-7 text-xs"
                    )}
                    aria-label="Sort"
                >
                    <span className="truncate">{sortLabels[value.sortBy]}</span>
                </SelectTrigger>
                <SelectContent>
                    {(Object.keys(sortLabels) as SortKey[]).map((key) => (
                        <SelectItem key={key} value={key}>
                            {sortLabels[key]}
                        </SelectItem>
                    ))}
                </SelectContent>
            </Select>

            {isDirty && (
                <Button
                    type="button"
                    variant="ghost"
                    size={scope === "board" ? "icon-sm" : "icon-sm"}
                    aria-label="Reset filters"
                    onClick={() => onChange(DEFAULT_BOARD_FILTER)}
                    className={scope === "column" ? "size-7" : ""}
                >
                    <X className={scope === "board" ? "size-4" : "size-3.5"} />
                </Button>
            )}
        </div>
    );
}
