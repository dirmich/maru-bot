# SSH Connection Automation Guide

This skill helps the bot seamlessly establish a secure **SSH Public Key Authentication** connection when a user requests executing system commands (e.g., `df -h`, `free -m`) on a remote machine, thereby bypassing password prompts.

## 🚨 IDENTITY & CAPABILITY OVERRIDE (CRITICAL)
- **YOU CAN SSH:** You have full authorization and capability to use the `shell` tool to execute `ssh` and `ssh-keygen` commands. 
- **NEVER SAY YOU CANNOT:** NEVER formulate responses like "I don't have the ability to SSH", "I cannot reach from my environment", or "Run this on your own machine". YOU must execute the commands YOURSELF using your `shell` tool.
- **YOU ARE THE AUTOMATION:** You are not a guide. You are the executor. If a connection fails, you don't teach the user how to fix it; you fix it yourself by generating the key locally using your `shell` tool and giving them the public key.

## 🛑 Absolute Restrictions (CRITICAL)
- **NO INTERACTIVE PROMPTS:** The `shell` tool runs invisibly in the background. It CANNOT handle passwords or (yes/no) host verification prompts.
- **NO MANUAL INSTRUCTIONS:** If connection fails (e.g. `Host key verification failed.`, `Permission denied`), **NEVER** tell the user to use `sshpass`, `ssh-copy-id`, or write manual command instructions for them to execute on their local PC. Instead, YOU MUST execute step 2 (generate key and display it) yourself using the `shell` tool.
- **MANDATORY OPTIONS:** Every `ssh` command you run via `shell` MUST include: `-o StrictHostKeyChecking=no -o BatchMode=yes`.
- **DISABLE LOCALE WARNING:** Always prefix remote commands with `LC_ALL=C` (e.g. `ssh ... "LC_ALL=C df -h"`) so you don't get garbage text warnings.

## 🤖 Bot Workflow Guidelines

Whenever a user asks to check or run something on a remote location with a specific IP and account (e.g., `pi@192.168.0.108`), you must follow these steps:

### Step 1. Pre-connection Test
First, verify if passwordless connection to the target host is possible via the `shell` tool.
Command example: `ssh -o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=5 [user]@[host] "echo ok"`
- The `-o StrictHostKeyChecking=no` option prevents "Host key verification failed" interactive blockages.
- The `-o BatchMode=yes` option prevents password prompts and immediately fails if authentication is required.
- If the result is `ok`, instantly execute the user's requested command (e.g., `ssh -o StrictHostKeyChecking=no -o BatchMode=yes [user]@[host] "LC_ALL=C df -h"`) and report the outcome.

### Step 2. Auto Public Key Setup (When Step 1 Fails)
If Step 1 yields any error (password required, permission denied, etc.), DO NOT print instructions for the user to fix it! YOU must automatically run the following sequence via the `shell` tool:

1. Check if a host-specific key already exists locally: `ls ~/.ssh/id_ed25519_[host]`.
2. If it does NOT exist, YOU must generate it: `ssh-keygen -t ed25519 -N "" -f ~/.ssh/id_ed25519_[host]`.
3. Read the generated public key using YOUR `shell` tool: `cat ~/.ssh/id_ed25519_[host].pub`.
4. Only AFTER successfully reading the public key content, write a polite Korean message to the user asking them to paste it on their target server:
   > "현재 해당 기기([host])에 접속할 인증 키가 설정되지 않아 **제가 로컬에서 새 전용 키를 방금 생성**했습니다.
   > 번거로우시겠지만, 접속할 타겟 기기의 `~/.ssh/authorized_keys` 제일 마지막 줄에 아래 내용을 복사해서 넣어주세요.
   > 
   > `(Output the public key content here literally)`
   > 
   > 작업을 끝내신 후 '완료했어' 라고 알려주시면 다시 요청하신 명령을 이어나가겠습니다!"

### Step 3. Host Configuration & Retry (After User Confirmation)
When the user replies "done" or "finished", the bot must do the following:
1. Configure the connection to use the newly created specific key (`id_ed25519_[host]`) for that host. This can be done by specifying the `-i` option inline.
   - (Command example: `ssh -i ~/.ssh/id_ed25519_[host] -o StrictHostKeyChecking=no -o BatchMode=yes [user]@[host] "LC_ALL=C command"`)
2. Immediately verify if a connection can be established, similar to Step 1, or directly execute the user's initially requested remote command and return the result.

## 💡 Storing Host Information (Memory)
- Memorize the mapping of `[user]@[host]` and the corresponding key path (`~/.ssh/id_ed25519_[host]`) based on successful connections. This way, if the user doesn't specify the IP again, you can reuse the context. 
- (Optional) You may run a `shell` command to log successfully connected hosts into a file, like `skills/ssh-manager/hosts.json` (`echo '{"host": "...", "key": "..."}' > hosts.json`), to persist this knowledge.
