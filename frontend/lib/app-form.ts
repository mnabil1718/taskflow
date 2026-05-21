import { createFormHook } from "@tanstack/react-form";
import { fieldContext, formContext } from "./form-context";
import { InputField } from "@/components/form/input-field";
import { TextareaField } from "@/components/form/textarea-field";
import { SelectField } from "@/components/form/select-field";
import { SubmitButton } from "@/components/form/submit-button";

export const { useAppForm, withForm } = createFormHook({
  fieldComponents: { InputField, TextareaField, SelectField },
  formComponents: { SubmitButton },
  fieldContext,
  formContext,
});
