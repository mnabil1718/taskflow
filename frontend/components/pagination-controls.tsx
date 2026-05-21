"use client";

import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
} from "@/components/ui/pagination";
import { cn } from "@/lib/utils";

interface PaginationControlsProps {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
    onPageChange: (page: number) => void;
    disabled?: boolean;
}

// Computes the page numbers (and ellipsis markers) to show in the bar.
// Keeps the first, last, current, and one neighbour each side; collapses
// the rest with ellipsis. Short lists (<=7 pages) render every page.
function buildPageList(page: number, totalPages: number): (number | "ellipsis-left" | "ellipsis-right")[] {
    if (totalPages <= 7) {
        return Array.from({ length: totalPages }, (_, i) => i + 1);
    }

    const pages: (number | "ellipsis-left" | "ellipsis-right")[] = [1];

    if (page > 3) pages.push("ellipsis-left");

    const start = Math.max(2, page - 1);
    const end = Math.min(totalPages - 1, page + 1);
    for (let i = start; i <= end; i++) pages.push(i);

    if (page < totalPages - 2) pages.push("ellipsis-right");

    pages.push(totalPages);
    return pages;
}

export function PaginationControls({
    page,
    pageSize,
    total,
    totalPages,
    onPageChange,
    disabled,
}: PaginationControlsProps) {
    if (totalPages <= 0) return null;

    const firstOnPage = (page - 1) * pageSize + 1;
    const lastOnPage = Math.min(page * pageSize, total);

    const handleSelect = (e: React.MouseEvent, target: number) => {
        e.preventDefault();
        if (disabled || target === page || target < 1 || target > totalPages) return;
        onPageChange(target);
    };

    const prevDisabled = disabled || page <= 1;
    const nextDisabled = disabled || page >= totalPages;
    const inactiveClass = (isDisabled: boolean) =>
        cn(isDisabled && "pointer-events-none opacity-50");

    const pageItems = buildPageList(page, totalPages);

    return (
        <div className="flex flex-col items-center justify-between gap-3 sm:flex-row">
            <p className="text-sm text-muted-foreground">
                Showing{" "}
                <span className="font-medium text-foreground tabular-nums">
                    {firstOnPage}
                </span>
                –
                <span className="font-medium text-foreground tabular-nums">
                    {lastOnPage}
                </span>{" "}
                of{" "}
                <span className="font-medium text-foreground tabular-nums">{total}</span>
                {" · "}
                <span className="tabular-nums">
                    Page {page} of {totalPages}
                </span>
            </p>

            <Pagination className="mx-0 w-auto justify-end">
                <PaginationContent>
                    <PaginationItem>
                        <PaginationPrevious
                            href="#"
                            aria-disabled={prevDisabled}
                            tabIndex={prevDisabled ? -1 : 0}
                            className={inactiveClass(prevDisabled)}
                            onClick={(e) => handleSelect(e, page - 1)}
                        />
                    </PaginationItem>

                    {pageItems.map((item, idx) =>
                        item === "ellipsis-left" || item === "ellipsis-right" ? (
                            <PaginationItem key={`${item}-${idx}`}>
                                <PaginationEllipsis />
                            </PaginationItem>
                        ) : (
                            <PaginationItem key={item}>
                                <PaginationLink
                                    href="#"
                                    isActive={item === page}
                                    aria-disabled={disabled}
                                    tabIndex={disabled ? -1 : 0}
                                    className={inactiveClass(!!disabled)}
                                    onClick={(e) => handleSelect(e, item)}
                                >
                                    {item}
                                </PaginationLink>
                            </PaginationItem>
                        )
                    )}

                    <PaginationItem>
                        <PaginationNext
                            href="#"
                            aria-disabled={nextDisabled}
                            tabIndex={nextDisabled ? -1 : 0}
                            className={inactiveClass(nextDisabled)}
                            onClick={(e) => handleSelect(e, page + 1)}
                        />
                    </PaginationItem>
                </PaginationContent>
            </Pagination>
        </div>
    );
}
