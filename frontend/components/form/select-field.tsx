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

export interface SelectOption {
    value: string;
    label: string;
    /** Secondary text rendered below the label in the dropdown (not shown in trigger). */
    description?: string;
}

interface SelectFieldProps {
    label: string;
    options: SelectOption[];
    placeholder?: string;
    required?: boolean;
    desc?: string;
    /** Called with the new value on every change, alongside the internal field update. */
    onChange?: (value: string) => void;
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
    onChange,
}: SelectFieldProps) {
    const field = useFieldContext<string>();
    const errors = field.state.meta.errors;
    const selectedOption = options.find((o) => o.value === field.state.value);

    return (
        <div className="space-y-1.5">
            <Label htmlFor={field.name}>
                <span>{label}{required && <span className="ml-0.5 text-destructive">*</span>}</span>
            </Label>
            <Select
                value={field.state.value}
                onValueChange={(value) => {
                    const v = value ?? "";
                    field.handleChange(v);
                    onChange?.(v);
                }}
            >
                <SelectTrigger
                    id={field.name}
                    className="w-full"
                    aria-invalid={errors.length > 0}
                >
                    {selectedOption ? (
                        <span data-slot="select-value" className="flex flex-1 text-left text-sm truncate">
                            {selectedOption.label}
                        </span>
                    ) : (
                        <SelectValue placeholder={placeholder} />
                    )}
                </SelectTrigger>
                <SelectContent>
                    {options.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value} description={opt.description}>
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
