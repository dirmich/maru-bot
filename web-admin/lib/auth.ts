import { NextAuthOptions } from "next-auth";
import GoogleProvider from "next-auth/providers/google";
import { PrismaAdapter } from "@auth/prisma-adapter";
import { PrismaClient } from "@prisma/client";

// Build-time safety for Prisma
const prisma = (function () {
    if (typeof window !== 'undefined' || process.env.NEXT_PHASE === 'phase-production-build') {
        return null;
    }
    return new PrismaClient();
})();

export const authOptions: NextAuthOptions = {
    adapter: prisma ? (PrismaAdapter(prisma) as any) : undefined,
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
