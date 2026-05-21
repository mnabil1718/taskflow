"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { config } from "@/lib/config";
import { tokenStorage } from "@/lib/token";

export type NotificationEventType =
    | "task.created"
    | "task.updated"
    | "task.deleted"
    | "task.assigned"
    | "task.moved"
    | "task.deadline_reminder";

export interface NotificationItem {
    id: string;
    type: NotificationEventType;
    taskId: string;
    projectId: string;
    taskTitle?: string;
    reminderWindow?: string;
    timestamp: string;
    read: boolean;
}

const MAX_ITEMS = 50;
const RECONNECT_BASE_MS = 2_000;
const RECONNECT_MAX_MS = 30_000;

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

export function useSseNotifications(enabled: boolean) {
    const [items, setItems] = useState<NotificationItem[]>([]);
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

                for (const { event, data } of parseSseBlocks(complete)) {
                    try {
                        const payload = JSON.parse(data);
                        const item: NotificationItem = {
                            id: `${payload.task_id}-${payload.timestamp}-${Math.random().toString(36).slice(2)}`,
                            type: event as NotificationEventType,
                            taskId: payload.task_id,
                            projectId: payload.project_id,
                            taskTitle: payload.task?.title,
                            reminderWindow: payload.reminder_window,
                            timestamp: payload.timestamp,
                            read: false,
                        };
                        setItems((prev) => [item, ...prev].slice(0, MAX_ITEMS));
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
    }, []);

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

    const markAllRead = useCallback(
        () => setItems((prev) => prev.map((n) => ({ ...n, read: true }))),
        []
    );

    const markRead = useCallback(
        (id: string) =>
            setItems((prev) =>
                prev.map((n) => (n.id === id ? { ...n, read: true } : n))
            ),
        []
    );

    const unreadCount = items.filter((n) => !n.read).length;

    return { notifications: items, unreadCount, markAllRead, markRead };
}
