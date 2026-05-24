import { Suspense } from "react";
import { RegisterForm } from "@/components/auth/register-form";

export const metadata = { title: "Register - TaskFlow" };

export default function RegisterPage() {
    return (
        <Suspense>
            <RegisterForm />
        </Suspense>
    );
}
