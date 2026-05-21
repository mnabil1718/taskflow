"use client";

import { useState } from "react";
import { Eye, EyeOff } from "lucide-react";
import { useFieldContext } from "@/lib/form-context";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface InputFieldProps {
    label: string;
    type?: "text" | "email" | "password" | "date";
    placeholder?: string;
    autoComplete?: string;
    required?: boolean;
    desc?: string;
    min?: string;
}

function getErrorMessage(error: unknown): string {
    if (typeof error === "object" && error !== null && "message" in error) {
        return String((error as { message: string }).message);
    }
    return String(error);
}

export function InputField({
    label,
    type = "text",
    placeholder,
    autoComplete,
    required,
    desc,
    min,
}: InputFieldProps) {
    const field = useFieldContext<string>();
    const errors = field.state.meta.errors;
    const [showPassword, setShowPassword] = useState(false);

    const isPassword = type === "password";
    const resolvedType = isPassword ? (showPassword ? "text" : "password") : type;

    return (
        <div className="space-y-1.5">
            <Label htmlFor={field.name}>
                <span>{label}{required && <span className="ml-0.5 text-destructive">*</span>}</span>
            </Label>
            <div className="relative">
                <Input
                    id={field.name}
                    name={field.name}
                    type={resolvedType}
                    placeholder={placeholder}
                    autoComplete={autoComplete}
                    min={min}
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    aria-invalid={errors.length > 0}
                    className={
                        cn("px-3! py-5!",
                            isPassword ? "pr-10!" : ""
                        )}
                />
                {isPassword && (
                    <button
                        type="button"
                        onClick={() => setShowPassword((v) => !v)}
                        className="absolute inset-y-0 right-0 flex items-center px-3 text-muted-foreground hover:text-foreground"
                        aria-label={showPassword ? "Hide password" : "Show password"}
                    >
                        {showPassword ? (
                            <EyeOff className="h-4 w-4" />
                        ) : (
                            <Eye className="h-4 w-4" />
                        )}
                    </button>
                )}
            </div>
            {errors.length > 0 && (
                <p className="text-[0.8rem] font-medium text-destructive">
                    {getErrorMessage(errors[0])}
                </p>
            )}
            {desc && errors.length == 0 && (
                <p className="text-[0.8rem] text-muted-foreground">
                    {desc}
                </p>
            )
            }
        </div>
    );
}
