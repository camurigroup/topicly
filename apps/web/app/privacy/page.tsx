import type { Metadata } from "next";

import { LegalPage } from "@/components/legal-page";

export const metadata: Metadata = { title: "Política de privacidade" };

export default function PrivacyPage() {
  return <LegalPage title="Política de privacidade" settingKey="legal.privacy" />;
}
