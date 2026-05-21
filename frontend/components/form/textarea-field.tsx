"use client";

import { useFieldContext } from "@/lib/form-context";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

interface TextareaFieldProps {
    label: string;
    placeholder?: string;
    required?: boolean;
    desc?: string;
    rows?: number;
}

function getErrorMessage(error: unknown): string {
    if (typeof error === "object" && error !== null && "message" in error) {
        return String((error as { message: string }).message);
    }
    return String(error);
}

export function TextareaField({
    label,
    placeholder,
    required,
    desc,
    rows = 3,
}: TextareaFieldProps) {
    const field = useFieldContext<string>();
    const errors = field.state.meta.errors;

    return (
        <div className="space-y-1.5">
            <Label htmlFor={field.name}>
                <span>{label}{required && <span className="ml-0.5 text-destructive">*</span>}</span>
            </Label>
            <Textarea
                id={field.name}
                name={field.name}
                placeholder={placeholder}
                rows={rows}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                aria-invalid={errors.length > 0}
            />
            {errors.length > 0 && (
                <p className="text-[0.8rem] font-medium text-destructive">
                    {getErrorMessage(errors[0])}
                </p>
            )}
            {desc && errors.length === 0 && (
                <p className="text-[0.8rem] text-muted-foreground">{desc}</p>
            )}
        </div>
    );
}
