# SSH Connection Automation Guide

This skill helps the bot seamlessly establish a secure **SSH Public Key Authentication** connection when a user requests executing system commands (e.g., `df -h`, `free -m`) on a remote machine, thereby bypassing password prompts.

## 🛑 Core Restrictions
- The provided `shell` tool runs in a background environment without a TTY, meaning it **cannot handle interactive password or host-key verification prompts**.
- **Crucial Rule:** If connection fails with errors like `Host key verification failed.` or `Permission denied`, DO NOT rely on your pre-trained knowledge to suggest tools like `sshpass`. You MUST strictly follow the 'Bot Workflow Guidelines' specified below (generating a key and asking the user to copy it).
- Always use `-o StrictHostKeyChecking=no` to bypass initial prompt blockages. Never blindly run an `ssh [user]@[host]` command without BatchMode.
- **Locale Warnings Avoidance:** To prevent the remote server from flooding the output with `bash: warning: setlocale: LC_ALL` errors which disrupts parsing, always prefix the remote command with `LC_ALL=C`. (e.g., `ssh ... "LC_ALL=C df -h"`).

## 🤖 Bot Workflow Guidelines

Whenever a user asks to check or run something on a remote location with a specific IP and account (e.g., `pi@192.168.0.108`), you must follow these steps:

### Step 1. Pre-connection Test
First, verify if passwordless connection to the target host is possible via the `shell` tool.
Command example: `ssh -o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=5 [user]@[host] "echo ok"`
- The `-o StrictHostKeyChecking=no` option prevents "Host key verification failed" interactive blockages.
- The `-o BatchMode=yes` option prevents password prompts and immediately fails if authentication is required.
- If the result is `ok`, instantly execute the user's requested command (e.g., `ssh -o StrictHostKeyChecking=no -o BatchMode=yes [user]@[host] "LC_ALL=C df -h"`) and report the outcome.

### Step 2. Public Key Setup (On Connection Failure)
If Step 1 fails, it means the local machine (the bot's current runtime environment, not the user's) lacks an authentication key, or the target machine does not recognize it.
Important: For security and management purposes, use separate, isolated keys for each host.

1. Check the `~/.ssh/` directory in the local environment.
2. Check if a host-specific key exists, e.g., `~/.ssh/id_ed25519_[host]`.
3. If it does not exist, use the `shell` tool to **generate a new dedicated key pair for that host** by running:
   - `ssh-keygen -t ed25519 -N "" -f ~/.ssh/id_ed25519_[host]` 
4. Read the contents of the generated or existing public key file `*.pub` (e.g., `id_ed25519_[host].pub`) using the `shell` tool.
5. Notify the user with the following message:
   > "I currently do not have passwordless access to the target machine ([host]). 
   > Please log in to your machine and append the following public key to the `~/.ssh/authorized_keys` file.
   > 
   > `(Output the parsed public key content here)`
   > 
   > Let me know when you are 'done' or 'finished', and I will retry the command!"

### Step 3. Host Configuration & Retry (After User Confirmation)
When the user replies "done" or "finished", the bot must do the following:
1. Configure the connection to use the newly created specific key (`id_ed25519_[host]`) for that host. This can be done by specifying the `-i` option inline.
   - (Command example: `ssh -i ~/.ssh/id_ed25519_[host] -o StrictHostKeyChecking=no -o BatchMode=yes [user]@[host] "LC_ALL=C command"`)
2. Immediately verify if a connection can be established, similar to Step 1, or directly execute the user's initially requested remote command and return the result.

## 💡 Storing Host Information (Memory)
- Memorize the mapping of `[user]@[host]` and the corresponding key path (`~/.ssh/id_ed25519_[host]`) based on successful connections. This way, if the user doesn't specify the IP again, you can reuse the context. 
- (Optional) You may run a `shell` command to log successfully connected hosts into a file, like `skills/ssh-manager/hosts.json` (`echo '{"host": "...", "key": "..."}' > hosts.json`), to persist this knowledge.
