"use client";

import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useState,
} from "react";
import { useRouter } from "next/navigation";
import { decodeJwt } from "jose";
import { authApi } from "./api/auth";
import { tokenStorage } from "./token";
import type { LoginRequest, RegisterRequest } from "./types";

interface AuthUser {
    id: string;
    email: string;
}

interface AuthContextValue {
    user: AuthUser | null;
    isLoading: boolean;
    login: (data: LoginRequest) => Promise<void>;
    register: (data: RegisterRequest) => Promise<void>;
    logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function userFromToken(token: string): AuthUser | null {
    try {
        const payload = decodeJwt(token);
        const id = payload["user_id"];
        const email = payload["email"];
        if (typeof id !== "string" || typeof email !== "string") return null;
        return { id, email };
    } catch {
        return null;
    }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<AuthUser | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const router = useRouter();

    useEffect(() => {
        const token = tokenStorage.getAccess();
        if (token) {
            setUser(userFromToken(token));
        }
        setIsLoading(false);
    }, []);

    const login = useCallback(async (data: LoginRequest) => {
        const tokens = await authApi.login(data);
        setUser(userFromToken(tokens.access_token));
    }, []);

    const register = useCallback(async (data: RegisterRequest) => {
        const tokens = await authApi.register(data);
        setUser(userFromToken(tokens.access_token));
    }, []);

    const logout = useCallback(async () => {
        await authApi.logout();
        setUser(null);
        router.push("/login");
    }, [router]);

    return (
        <AuthContext.Provider value={{ user, isLoading, login, register, logout }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth(): AuthContextValue {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error("useAuth must be used within AuthProvider");
    return ctx;
}
