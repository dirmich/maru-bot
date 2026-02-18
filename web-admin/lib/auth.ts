import { NextAuthOptions } from "next-auth";
import GoogleProvider from "next-auth/providers/google";
import { DrizzleAdapter } from "@auth/drizzle-adapter";
import { getDb } from "./db";

export const authOptions: NextAuthOptions = {
    adapter: (function () {
        const db = getDb();
        return db ? (DrizzleAdapter(db) as any) : undefined;
    })(),
    providers: [
        GoogleProvider({
            clientId: process.env.GOOGLE_CLIENT_ID || "dummy",
            clientSecret: process.env.GOOGLE_CLIENT_SECRET || "dummy",
        }),
    ],
    callbacks: {
        async signIn({ user }) {
            const adminEmail = process.env.ADMIN_GMAIL;
            if (!adminEmail) return true;
            return user.email === adminEmail;
        },
        async session({ session, user }) {
            if (session.user && user) {
                (session.user as any).id = user.id;
            }
            return session;
        },
    },
    pages: {
        signIn: '/auth/signin',
    },
    secret: process.env.NEXTAUTH_SECRET || "secret",
};
