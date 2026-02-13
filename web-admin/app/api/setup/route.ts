import { NextResponse } from "next/server";
import fs from "fs";
import path from "path";

export async function POST(req: Request) {
    try {
        const data = await req.json();
        const envPath = path.join(process.cwd(), ".env");

        let envContent = fs.readFileSync(envPath, "utf-8");

        // Simple regex replacement for keys
        const updates = {
            ADMIN_GMAIL: data.adminGmail,
            GOOGLE_CLIENT_ID: data.clientId,
            GOOGLE_CLIENT_SECRET: data.clientSecret,
            NEXTAUTH_SECRET: data.nextAuthSecret,
        };

        for (const [key, value] of Object.entries(updates)) {
            const regex = new RegExp(`^${key}=.*`, "m");
            if (envContent.match(regex)) {
                envContent = envContent.replace(regex, `${key}="${value}"`);
            } else {
                envContent += `\n${key}="${value}"`;
            }
        }

        fs.writeFileSync(envPath, envContent, "utf-8");

        return NextResponse.json({ success: true });
    } catch (e) {
        return NextResponse.json({ error: "Failed to save setup" }, { status: 500 });
    }
}
