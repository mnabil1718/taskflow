"use client";

import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

// Structural type covering the bits of a useAppForm() instance that the
// dialog needs to drive: the AppForm provider, a SubmitButton with the
// shared loading state, a Subscribe primitive for the cancel button's
// disabled state, and handleSubmit. Typed structurally so any
// useAppForm-returned instance fits without explicit generic plumbing.
interface FormDialogApi {
    AppForm: React.ComponentType<{ children?: React.ReactNode }>;
    SubmitButton: React.ComponentType<{
        children: React.ReactNode;
        className?: string;
    }>;
    Subscribe: React.ComponentType<{
        selector: (state: { isSubmitting: boolean }) => boolean;
        children: (isSubmitting: boolean) => React.ReactNode;
    }>;
    handleSubmit: () => void | Promise<void>;
}

interface FormDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    title: string;
    description?: string;
    /** Optional element used as the dialog trigger. Omit for externally-controlled dialogs. */
    trigger?: React.ReactElement;
    /** A useAppForm() instance whose state drives submission and the button labels. */
    form: FormDialogApi;
    submitLabel?: string;
    cancelLabel?: string;
    contentClassName?: string;
    children: React.ReactNode;
}

export function FormDialog({
    open,
    onOpenChange,
    title,
    description,
    trigger,
    form,
    submitLabel = "Save",
    cancelLabel = "Cancel",
    contentClassName,
    children,
}: FormDialogProps) {
    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            {trigger && <DialogTrigger render={trigger} />}
            <DialogContent className={cn("sm:max-w-md", contentClassName)}>
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                    {description && <DialogDescription>{description}</DialogDescription>}
                </DialogHeader>

                <form.AppForm>
                    <form
                        onSubmit={(e) => {
                            e.preventDefault();
                            form.handleSubmit();
                        }}
                        className="space-y-4"
                    >
                        {children}

                        <DialogFooter className="flex items-center gap-2">
                            <form.Subscribe selector={(state) => state.isSubmitting}>
                                {(isSubmitting) => (
                                    <Button
                                        type="button"
                                        size="lg"
                                        variant="outline"
                                        onClick={() => onOpenChange(false)}
                                        disabled={isSubmitting}
                                    >
                                        {cancelLabel}
                                    </Button>
                                )}
                            </form.Subscribe>
                            <form.SubmitButton className="sm:w-auto">
                                {submitLabel}
                            </form.SubmitButton>
                        </DialogFooter>
                    </form>
                </form.AppForm>
            </DialogContent>
        </Dialog>
    );
}
