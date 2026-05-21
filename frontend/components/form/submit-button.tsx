"use client";

import { useFormContext } from "@/lib/form-context";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";

interface SubmitButtonProps {
    children: React.ReactNode;
    className?: string;
}

export function SubmitButton({ children, className }: SubmitButtonProps) {
    const form = useFormContext();

    return (
        <form.Subscribe
            selector={(state) => state.isSubmitting}
            children={(isSubmitting) => (
                <Button
                    type="submit"
                    size={"lg"}
                    className={className ?? "w-full"}
                    disabled={isSubmitting}
                >
                    {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    {children}
                </Button>
            )}
        />
    );
}
