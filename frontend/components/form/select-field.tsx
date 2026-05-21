"use client";

import { useFieldContext } from "@/lib/form-context";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

interface SelectOption {
    value: string;
    label: string;
}

interface SelectFieldProps {
    label: string;
    options: SelectOption[];
    placeholder?: string;
    required?: boolean;
    desc?: string;
}

function getErrorMessage(error: unknown): string {
    if (typeof error === "object" && error !== null && "message" in error) {
        return String((error as { message: string }).message);
    }
    return String(error);
}

export function SelectField({
    label,
    options,
    placeholder,
    required,
    desc,
}: SelectFieldProps) {
    const field = useFieldContext<string>();
    const errors = field.state.meta.errors;

    return (
        <div className="space-y-1.5">
            <Label htmlFor={field.name}>
                <span>{label}{required && <span className="ml-0.5 text-destructive">*</span>}</span>
            </Label>
            <Select
                value={field.state.value}
                onValueChange={(value) => field.handleChange(value as string)}
            >
                <SelectTrigger
                    id={field.name}
                    className="w-full"
                    aria-invalid={errors.length > 0}
                >
                    <SelectValue placeholder={placeholder} />
                </SelectTrigger>
                <SelectContent>
                    {options.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                            {opt.label}
                        </SelectItem>
                    ))}
                </SelectContent>
            </Select>
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
