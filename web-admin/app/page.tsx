export const dynamic = "force-dynamic";

import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";
import { redirect } from "next/navigation";
import fs from "fs";
import path from "path";

export default async function IndexPage() {
  const session = await getServerSession(authOptions);

  // Check if .env is missing critical setup
  const envPath = path.join(process.cwd(), ".env");
  const envContent = fs.existsSync(envPath) ? fs.readFileSync(envPath, "utf-8") : "";
  const isSetupDone = envContent.includes('GOOGLE_CLIENT_ID="') && !envContent.includes('GOOGLE_CLIENT_ID=""');

  if (!isSetupDone) {
    redirect("/setup");
  }

  if (!session) {
    redirect("/auth/signin");
  }

  redirect("/chat");
}
