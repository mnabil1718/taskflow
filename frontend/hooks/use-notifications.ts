"use client";

import { useCallback, useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { config } from "@/lib/config";
import { tokenStorage } from "@/lib/token";
import { notificationsApi } from "@/lib/api/notifications";
import type { Notification, NotificationPage } from "@/lib/types";

const QUERY_KEY = ["notifications"] as const;
const RECONNECT_BASE_MS = 2_000;
const RECONNECT_MAX_MS = 30_000;
const EMPTY_PAGE: NotificationPage = { items: [], unread_count: 0 };

function parseSseBlocks(chunk: string): Array<{ event: string; data: string }> {
    const out: Array<{ event: string; data: string }> = [];
    for (const block of chunk.split("\n\n")) {
        let event = "";
        let data = "";
        for (const line of block.split("\n")) {
            if (line.startsWith("event: ")) event = line.slice(7).trim();
            else if (line.startsWith("data: ")) data = line.slice(6).trim();
        }
        if (event && data) out.push({ event, data });
    }
    return out;
}

// useNotifications backs the bell with the persisted feed: the REST list is
// the source of truth (so notifications survive refreshes and `read` state is
// stored server-side), and the SSE stream merges newly-created items into the
// same React Query cache by id. Because the live payload is the same shape as
// a listed item, a streamed notification dedupes cleanly against the fetched
// list instead of producing a transient duplicate.
export function useNotifications(enabled: boolean) {
    const qc = useQueryClient();

    const query = useQuery({
        queryKey: QUERY_KEY,
        queryFn: () => notificationsApi.list(),
        enabled,
        staleTime: 30_000,
    });

    // ingest folds one live notification into the cached page: skip if its id
    // is already present, otherwise prepend and bump the unread counter.
    const ingest = useCallback(
        (n: Notification) => {
            qc.setQueryData<NotificationPage>(QUERY_KEY, (prev) => {
                const page = prev ?? EMPTY_PAGE;
                if (page.items.some((it) => it.id === n.id)) return page;
                return {
                    items: [n, ...page.items],
                    unread_count: page.unread_count + (n.read_at ? 0 : 1),
                };
            });
        },
        [qc]
    );

    const abortRef = useRef<AbortController | null>(null);
    const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const delayRef = useRef(RECONNECT_BASE_MS);
    const deadRef = useRef(false);

    const connect = useCallback(async () => {
        if (deadRef.current) return;
        const token = tokenStorage.getAccess();
        if (!token) return;

        const ctrl = new AbortController();
        abortRef.current = ctrl;

        try {
            const res = await fetch(`${config.api.baseUrl}/notifications/stream`, {
                headers: { Authorization: `Bearer ${token}` },
                signal: ctrl.signal,
            });

            if (!res.ok || !res.body) throw new Error(`HTTP ${res.status}`);

            delayRef.current = RECONNECT_BASE_MS;
            const reader = res.body.getReader();
            const decoder = new TextDecoder();
            let buf = "";

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                buf += decoder.decode(value, { stream: true });

                const boundary = buf.lastIndexOf("\n\n");
                if (boundary === -1) continue;

                const complete = buf.slice(0, boundary + 2);
                buf = buf.slice(boundary + 2);

                for (const { data } of parseSseBlocks(complete)) {
                    try {
                        ingest(JSON.parse(data) as Notification);
                    } catch {
                        // skip malformed event
                    }
                }
            }
        } catch (err) {
            if (err instanceof DOMException && err.name === "AbortError") return;
        }

        if (!deadRef.current) {
            const delay = delayRef.current;
            delayRef.current = Math.min(delay * 2, RECONNECT_MAX_MS);
            retryRef.current = setTimeout(connect, delay);
        }
    }, [ingest]);

    useEffect(() => {
        if (!enabled) return;
        deadRef.current = false;
        connect();
        return () => {
            deadRef.current = true;
            abortRef.current?.abort();
            if (retryRef.current) clearTimeout(retryRef.current);
        };
    }, [enabled, connect]);

    const markAll = useMutation({
        mutationFn: () => notificationsApi.markAllRead(),
        onMutate: async () => {
            await qc.cancelQueries({ queryKey: QUERY_KEY });
            const prev = qc.getQueryData<NotificationPage>(QUERY_KEY);
            const now = new Date().toISOString();
            qc.setQueryData<NotificationPage>(QUERY_KEY, (p) => {
                const page = p ?? EMPTY_PAGE;
                return {
                    items: page.items.map((it) => (it.read_at ? it : { ...it, read_at: now })),
                    unread_count: 0,
                };
            });
            return { prev };
        },
        onError: (_e, _v, ctx) => {
            if (ctx?.prev) qc.setQueryData(QUERY_KEY, ctx.prev);
        },
    });

    const markOne = useMutation({
        mutationFn: (id: string) => notificationsApi.markRead(id),
        onMutate: async (id) => {
            await qc.cancelQueries({ queryKey: QUERY_KEY });
            const prev = qc.getQueryData<NotificationPage>(QUERY_KEY);
            const now = new Date().toISOString();
            qc.setQueryData<NotificationPage>(QUERY_KEY, (p) => {
                const page = p ?? EMPTY_PAGE;
                let dec = 0;
                const items = page.items.map((it) => {
                    if (it.id === id && !it.read_at) {
                        dec = 1;
                        return { ...it, read_at: now };
                    }
                    return it;
                });
                return { items, unread_count: Math.max(0, page.unread_count - dec) };
            });
            return { prev };
        },
        onError: (_e, _v, ctx) => {
            if (ctx?.prev) qc.setQueryData(QUERY_KEY, ctx.prev);
        },
    });

    const page = query.data ?? EMPTY_PAGE;

    return {
        notifications: page.items,
        unreadCount: page.unread_count,
        markAllRead: () => markAll.mutate(),
        markRead: (id: string) => markOne.mutate(id),
    };
}
