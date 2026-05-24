"use client";

import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";

import { useAppForm } from "@/lib/app-form";
import { useAuth } from "@/lib/auth-context";
import { loginSchema } from "@/schemas/login.schema";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Brand } from "../brand";

// safeRedirectTarget gates the `from` query param to internal paths only —
// an `?from=https://evil.example` would otherwise let the login form be
// used as an open-redirect oracle. We accept the param only if it's a
// site-relative path (starts with "/") that doesn't begin with the auth
// routes that would just bounce the user back here.
function safeRedirectTarget(from: string | null): string {
    if (!from || from === "/") return "/dashboard";
    if (!from.startsWith("/") || from.startsWith("//")) return "/dashboard";
    if (from.startsWith("/login") || from.startsWith("/register")) return "/dashboard";
    return from;
}

export function LoginForm() {
    const { login } = useAuth();
    const router = useRouter();
    const searchParams = useSearchParams();

    const form = useAppForm({
        defaultValues: { email: "", password: "" },
        onSubmit: async ({ value }) => {
            try {
                await login(value);
                const q = searchParams.get("from");
                const t = safeRedirectTarget(q);
                router.push(t);
            } catch {
                // error is toasted by the api interceptor
            }
        },
    });

    return (
        <Card className="w-full max-w-md p-7!">
            <div className="flex w-full justify-center mb-3">
                <Brand />
            </div>
            <CardHeader className="pb-2">
                <CardTitle className="text-2xl font-semibold tracking-tight">
                    Sign in to your account
                </CardTitle>
                <CardDescription>
                    Enter your credentials to access TaskFlow
                </CardDescription>
            </CardHeader>

            <CardContent className="pt-6">
                <form.AppForm>
                    <form
                        onSubmit={(e) => {
                            e.preventDefault();
                            form.handleSubmit();
                        }}
                        className="space-y-5"
                    >
                        <form.AppField
                            name="email"
                            validators={{
                                onChange: loginSchema.shape.email,
                                onBlur: loginSchema.shape.email,
                            }}
                            children={(field) => (
                                <field.InputField
                                    label="Email"
                                    type="email"
                                    placeholder="you@example.com"
                                    autoComplete="email"
                                />
                            )}
                        />

                        <form.AppField
                            name="password"
                            validators={{
                                onChange: loginSchema.shape.password,
                                onBlur: loginSchema.shape.password,
                            }}
                            children={(field) => (
                                <field.InputField
                                    label="Password"
                                    type="password"
                                    autoComplete="current-password"
                                    placeholder="Enter your password"
                                />
                            )}
                        />

                        <div className="pt-2 space-y-3">
                            <form.SubmitButton>Sign in</form.SubmitButton>

                            <p className="text-center text-sm text-muted-foreground">
                                Don&apos;t have an account?{" "}
                                <Link
                                    href="/register"
                                    className="text-foreground underline hover:text-foreground/80"
                                >
                                    Register
                                </Link>
                            </p>
                        </div>
                    </form>
                </form.AppForm>
            </CardContent>
        </Card>
    );
}
