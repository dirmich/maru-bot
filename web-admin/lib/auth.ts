import { NextAuthOptions } from "next-auth";
import GoogleProvider from "next-auth/providers/google";
import { PrismaAdapter } from "@auth/prisma-adapter";
import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

export const authOptions: NextAuthOptions = {
    adapter: PrismaAdapter(prisma) as any,
    providers: [
        GoogleProvider({
            clientId: process.env.GOOGLE_CLIENT_ID || "",
            clientSecret: process.env.GOOGLE_CLIENT_SECRET || "",
        }),
    ],
    callbacks: {
        async signIn({ user }) {
            const adminEmail = process.env.ADMIN_GMAIL;
            if (!adminEmail) return true; // Allow all if not set yet (first setup)
            return user.email === adminEmail;
        },
        async session({ session, user }) {
            if (session.user) {
                (session.user as any).id = user.id;
            }
            return session;
        },
    },
    pages: {
        signIn: '/auth/signin',
    },
};
