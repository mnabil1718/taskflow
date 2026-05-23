"use client";

import { useEffect, useRef, useState } from "react";
import { CheckIcon, ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";

interface Option {
    value: string;
    label: string;
}

interface FilterMultiSelectProps {
    options: Option[];
    value: string[];
    onChange: (value: string[]) => void;
    placeholder: string;
    className?: string;
}

export function FilterMultiSelect({
    options,
    value,
    onChange,
    placeholder,
    className,
}: FilterMultiSelectProps) {
    const [open, setOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
                setOpen(false);
            }
        };
        if (open) document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, [open]);

    const toggle = (v: string) => {
        onChange(value.includes(v) ? value.filter((x) => x !== v) : [...value, v]);
    };

    const label =
        value.length === 0
            ? placeholder
            : value.length === 1
            ? (options.find((o) => o.value === value[0])?.label ?? value[0])
            : `${value.length} selected`;

    return (
        <div ref={containerRef} className={cn("relative", className)}>
            <button
                type="button"
                onClick={() => setOpen((o) => !o)}
                className={cn(
                    "flex h-8 w-fit items-center justify-between gap-1.5 rounded-lg border border-input bg-transparent py-2 pr-2 pl-2.5 text-sm whitespace-nowrap transition-colors outline-none select-none",
                    "hover:bg-accent/30 focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50",
                    "dark:bg-input/30 dark:hover:bg-input/50",
                    value.length === 0 && "text-muted-foreground"
                )}
            >
                <span className="flex flex-1 text-left text-sm truncate">{label}</span>
                <ChevronDown className="pointer-events-none size-4 text-muted-foreground" />
            </button>

            {open && (
                <div className="absolute z-50 mt-1 min-w-36 overflow-hidden rounded-lg bg-popover text-popover-foreground shadow-md ring-1 ring-foreground/10">
                    <ul className="p-1">
                        {options.map((option) => {
                            const checked = value.includes(option.value);
                            return (
                                <li key={option.value}>
                                    <button
                                        type="button"
                                        onClick={() => toggle(option.value)}
                                        className={cn(
                                            "relative flex w-full cursor-default items-center rounded-md py-1 pr-8 pl-1.5 text-sm outline-hidden select-none",
                                            "hover:bg-accent hover:text-accent-foreground"
                                        )}
                                    >
                                        <span className="flex-1 text-left whitespace-nowrap">{option.label}</span>
                                        {checked && (
                                            <CheckIcon className="pointer-events-none absolute right-2 size-4" />
                                        )}
                                    </button>
                                </li>
                            );
                        })}
                    </ul>
                </div>
            )}
        </div>
    );
}
