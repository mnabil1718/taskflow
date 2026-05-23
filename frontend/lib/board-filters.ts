import { compareLexorank } from "./lexorank";
import type { Task, TaskPriority } from "./types";

// SortKey enumerates how a column can order its tasks. "position" is the
// drag-and-drop manual ordering (lexorank string); the others are derived
// orderings the user can pick from the toolbar.
export type SortKey = "position" | "title" | "deadline" | "priority";

export interface BoardFilter {
    search: string;
    sortBy: SortKey;
}

export const DEFAULT_BOARD_FILTER: BoardFilter = {
    search: "",
    sortBy: "position",
};

// Priority ordering used by the priority sort. Higher number = higher
// importance so the .sort callback flips this to descending naturally.
const priorityWeight: Record<TaskPriority, number> = {
    high: 3,
    medium: 2,
    low: 1,
};

// composeFilters returns the *effective* filter for a column: a column's
// own values override the board-wide defaults. Empty search falls through
// to the board; sortBy always uses the column's value (the toolbar always
// renders one — there is no "unset" state). Calling this is idempotent —
// same inputs always produce the same output — so callers can recompute
// it in every render without behavioural drift.
export function composeFilters(
    board: BoardFilter,
    column: BoardFilter
): BoardFilter {
    return {
        search: column.search.trim() || board.search.trim(),
        sortBy: column.sortBy === "position" && board.sortBy !== "position"
            ? board.sortBy
            : column.sortBy,
    };
}

// applyFilter is the pure data transformation each column runs over its
// task list: case-insensitive title + assignee match, then a stable sort.
// All non-"position" sorts use lexorank as a final tiebreaker so equal
// keys still produce a deterministic order — important so the rendered
// list doesn't shuffle between renders.
export function applyFilter(tasks: Task[], filter: BoardFilter): Task[] {
    const query = filter.search.trim().toLowerCase();

    const filtered = query
        ? tasks.filter((t) => {
              if (t.title.toLowerCase().includes(query)) return true;
              if (t.assignee_name?.toLowerCase().includes(query)) return true;
              if (t.assignee_email?.toLowerCase().includes(query)) return true;
              return false;
          })
        : tasks;

    const sorted = [...filtered];
    switch (filter.sortBy) {
        case "position":
            sorted.sort((a, b) => compareLexorank(a.position, b.position));
            break;
        case "title":
            sorted.sort(
                (a, b) =>
                    a.title.localeCompare(b.title) ||
                    compareLexorank(a.position, b.position)
            );
            break;
        case "deadline":
            sorted.sort((a, b) => {
                // Tasks without a due date sink to the bottom so the
                // user's attention starts at imminent deadlines.
                if (!a.due_date && !b.due_date) return compareLexorank(a.position, b.position);
                if (!a.due_date) return 1;
                if (!b.due_date) return -1;
                const diff = new Date(a.due_date).getTime() - new Date(b.due_date).getTime();
                return diff !== 0 ? diff : compareLexorank(a.position, b.position);
            });
            break;
        case "priority":
            sorted.sort((a, b) => {
                const diff = priorityWeight[b.priority] - priorityWeight[a.priority];
                return diff !== 0 ? diff : compareLexorank(a.position, b.position);
            });
            break;
    }
    return sorted;
}
