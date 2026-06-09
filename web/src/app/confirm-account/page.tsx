import { Suspense } from "react";
import ConfirmAccountForm from "./ConfirmAccountForm";

export default function ConfirmAccountPage() {
  return (
    <Suspense fallback={<div style={{ minHeight: "100vh" }} />}>
      <ConfirmAccountForm />
    </Suspense>
  );
}
