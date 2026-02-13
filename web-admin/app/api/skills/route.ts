export const dynamic = "force-dynamic";

import { NextResponse } from 'next/server';
import { exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';

const execAsync = promisify(exec);

// Path to the maruminibot binary or go run command
const MARUBOT_CMD = process.platform === 'win32' ? 'go run ..\\cmd\\maruminibot\\main.go' : 'maruminibot';

export async function GET() {
    try {
        // We want JSON output but the CLI might not support it yet.
        // Let's check if we can get it or just parse the text.
        // For now, I'll list the files in the skills directory as a fallback.
        const { stdout } = await execAsync(`${MARUBOT_CMD} skills list`, {
            cwd: path.join(process.cwd(), '..')
        });

        return NextResponse.json({ output: stdout });
    } catch (error: any) {
        return NextResponse.json({ error: error.message }, { status: 500 });
    }
}

export async function POST(request: Request) {
    try {
        const { action, skill } = await request.json();
        let command = '';
        if (action === 'install') {
            command = `${MARUBOT_CMD} skills install ${skill}`;
        } else if (action === 'remove') {
            command = `${MARUBOT_CMD} skills remove ${skill}`;
        } else {
            return NextResponse.json({ error: 'Invalid action' }, { status: 400 });
        }

        const { stdout, stderr } = await execAsync(command, {
            cwd: path.join(process.cwd(), '..')
        });

        return NextResponse.json({ stdout, stderr });
    } catch (error: any) {
        return NextResponse.json({ error: error.message }, { status: 500 });
    }
}
