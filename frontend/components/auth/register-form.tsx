"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";

import { useAppForm } from "@/lib/app-form";
import { useAuth } from "@/lib/auth-context";
import { ApiError } from "@/lib/api";
import { registerSchema } from "@/schemas/register.schema";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Brand } from "@/components/brand";

export function RegisterForm() {
    const { register } = useAuth();
    const router = useRouter();

    const form = useAppForm({
        defaultValues: { name: "", email: "", password: "", confirm_password: "" },
        onSubmit: async ({ value }) => {
            try {
                await register({
                    name: value.name,
                    email: value.email,
                    password: value.password,
                });
                router.push("/projects");
            } catch (err) {
                toast.error(
                    err instanceof ApiError
                        ? err.message
                        : "Something went wrong. Please try again."
                );
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
                    Create an account
                </CardTitle>
                <CardDescription>
                    Sign up to start managing your tasks
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
                            name="name"
                            validators={{
                                onChange: registerSchema.shape.name,
                                onBlur: registerSchema.shape.name,
                            }}
                            children={(field) => (
                                <field.InputField
                                    label="Name"
                                    type="text"
                                    placeholder="John Doe"
                                    autoComplete="name"
                                    required
                                />
                            )}
                        />

                        <form.AppField
                            name="email"
                            validators={{
                                onChange: registerSchema.shape.email,
                                onBlur: registerSchema.shape.email,
                            }}
                            children={(field) => (
                                <field.InputField
                                    label="Email"
                                    type="email"
                                    placeholder="you@example.com"
                                    autoComplete="email"
                                    required
                                />
                            )}
                        />

                        <form.AppField
                            name="password"
                            validators={{
                                onChange: registerSchema.shape.password,
                                onBlur: registerSchema.shape.password,
                            }}
                            children={(field) => (
                                <field.InputField
                                    label="Password"
                                    type="password"
                                    placeholder="Enter new password"
                                    autoComplete="new-password"
                                    desc="Must be at least 8 characters"
                                    required
                                />
                            )}
                        />

                        <form.AppField
                            name="confirm_password"
                            validators={{
                                onBlur: registerSchema.shape.confirm_password,
                            }}
                            children={(field) => (
                                <field.InputField
                                    label="Confirm password"
                                    type="password"
                                    placeholder="Repeat your password"
                                    autoComplete="new-password"
                                    required
                                />
                            )}
                        />

                        <div className="pt-2 space-y-3">
                            <form.SubmitButton>Create account</form.SubmitButton>

                            <p className="text-center text-sm text-muted-foreground">
                                Already have an account?{" "}
                                <Link
                                    href="/login"
                                    className="text-foreground underline hover:text-foreground/80"
                                >
                                    Sign in
                                </Link>
                            </p>
                        </div>
                    </form>
                </form.AppForm>
            </CardContent>
        </Card>
    );
}
