export const dynamic = "force-dynamic";

import { NextResponse } from 'next/server';
import { addMessage, getMessages } from '@/lib/db';
import { exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';

const execAsync = promisify(exec);
const MARUBOT_CMD = process.platform === 'win32' ? 'go run ..\\cmd\\marubot\\main.go' : 'marubot';

export async function GET() {
    const messages = await getMessages();
    return NextResponse.json(messages.reverse());
}

export async function POST(request: Request) {
    try {
        const { message } = await request.json();

        // Save user message
        await addMessage('user', message);

        // Call marubot agent
        // We use a single message mode
        const { stdout, stderr } = await execAsync(`${MARUBOT_CMD} agent -m "${message.replace(/"/g, '\\"')}"`, {
            cwd: path.join(process.cwd(), '..')
        });

        const response = stdout.trim();

        // The output might contain the logo 🦞 and some prefix, let's clean it up if needed
        // In main.go: fmt.Printf("\n%s %s\n", logo, response)
        const cleanedResponse = response.replace(/^.*?🦞\s*/, '').trim();

        // Save assistant message
        await addMessage('assistant', cleanedResponse);

        return NextResponse.json({ response: cleanedResponse });
    } catch (error: any) {
        console.error('Chat error:', error);
        return NextResponse.json({ error: error.message }, { status: 500 });
    }
}
