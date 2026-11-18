import type { Metadata } from "next";

import { LegalPage } from "@/components/legal-page";

export const metadata: Metadata = { title: "Termos de uso" };

export default function TermsPage() {
  return <LegalPage title="Termos de uso" settingKey="legal.terms" />;
}
