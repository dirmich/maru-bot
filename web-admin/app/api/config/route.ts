export const dynamic = "force-dynamic";

import { NextResponse } from 'next/server';
import { getConfig, saveConfig } from '@/lib/config';

export async function GET() {
    const config = getConfig();
    if (!config) {
        return NextResponse.json({ error: 'Config not found' }, { status: 404 });
    }
    return NextResponse.json(config);
}

export async function POST(request: Request) {
    try {
        const config = await request.json();
        saveConfig(config);
        return NextResponse.json({ success: true });
    } catch (error) {
        return NextResponse.json({ error: 'Failed to save config' }, { status: 500 });
    }
}
