import fs from 'fs';
import path from 'path';
import os from 'os';

const CONFIG_PATH = path.join(os.homedir(), '.marubot', 'config.json');

export function getConfig() {
    if (!fs.existsSync(CONFIG_PATH)) {
        return null;
    }
    const data = fs.readFileSync(CONFIG_PATH, 'utf-8');
    return JSON.parse(data);
}

export function saveConfig(config: any) {
    const dir = path.dirname(CONFIG_PATH);
    if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
    }
    fs.writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2), 'utf-8');
}
