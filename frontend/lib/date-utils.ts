// Formats a date as "21 Feb 2026". Accepts an ISO string or a Date so it can
// be reused wherever the app needs a compact, locale-agnostic date label.
export function formatDate(input: string | Date): string {
    const d = typeof input === "string" ? new Date(input) : input;
    return new Intl.DateTimeFormat("en-GB", {
        day: "numeric",
        month: "short",
        year: "numeric",
    }).format(d);
}

export function formatDistanceToNow(iso: string): string {
    const diff = Date.now() - new Date(iso).getTime();
    const s = Math.floor(diff / 1000);
    if (s < 60) return "just now";
    const m = Math.floor(s / 60);
    if (m < 60) return `${m}m ago`;
    const h = Math.floor(m / 60);
    if (h < 24) return `${h}h ago`;
    const d = Math.floor(h / 24);
    if (d < 30) return `${d}d ago`;
    return new Date(iso).toLocaleDateString();
}
