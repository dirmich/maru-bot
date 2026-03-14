<<<<<<< HEAD
# MaruBot - Ultra-lightweight AI Assistant (v0.4.60) 🦞
=======
# MaruBot - Ultra-lightweight AI Assistant (v0.4.61) 🦞
>>>>>>> d992838 (Release v0.4.62: Refactor config structure and fix vLLM model name issue)

**MaruBot** is an ultra-efficient, physical AI assistant inspired by the philosophy of [PicoClaw](https://github.com/sipeed/picoclaw). It is designed to run in extremely constrained environments, requiring as little as **10MB of RAM**. MaruBot supports Raspberry Pi, Generic Linux servers, and Windows PCs, and features a powerful 'Auto-Evolution' engine that allows it to expand its own capabilities autonomously.

---

## ✨ Key Features

### 1. 🚀 Ultra-lightweight & High Performance
- **10MB RAM Usage:** Optimized Go binary ensures smooth operation even on low-end embedded devices.
- **Single Binary:** No-dependency deployment makes installation and management effortless.

### 2. 🌍 Multi-Platform Support
- **Raspberry Pi:** Full support for GPIO, camera, and hardware sensors (ARM32/64).
- **Linux:** Compatible with Ubuntu, Debian, AWS EC2, and other standard Linux distros.
- **Windows:** Official support for both 64-bit (amd64) and 32-bit (386) architectures.
- **macOS:** Official support for Intel and Apple Silicon (arm64) via DMG installers.

### 3. 🧬 Auto-Evolution Engine
- **`create_tool`**: The agent can autonomously code and register new atomic tools using Bash or Python scripts.
- **`create_skill`**: Automatically generates high-level 'Skills' (folder-based guidelines and workflows) to master new domains.

### 4. 🧠 Smart Memory System (RAG)
- **SQLite FTS5:** Persistent long-term memory using SQLite Full-Text Search for lightning-fast context retrieval.
- **Facts & Preferences:** Remembers user preferences and core facts across sessions.

### 5. 🛠️ Advanced Automation & Hardware Integration
- **Cron Scheduler:** Supports scheduled tasks like "Remind me of the weather at 9 AM tomorrow."
- **SSH Manager:** Intelligent management of SSH keys and remote server connections.
- **MAVLink Drone:** Capability to interface with flight controllers for physical drone missions.

---

## 📂 Project Structure
- `/config`: Core hardware and agent configuration files.
- `/skills`: Specialized knowledge and domain guidelines (`SKILL.md`).
- `/extensions`: Dynamic tool scripts created via `create_tool`.
- `/memory`: SQLite conversation database and fact storage.

---

## 🚀 Quick Start

### 1. 🐧 Linux / 🍎 macOS (Terminal)
Run the following command in your terminal to install instantly (requires curl):
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. 🪟 Windows (GUI/Manual)
For Windows users, we recommend downloading the files from the **Official Release Page** instead of using terminal commands:
1. Visit the [Official Release Page](https://github.com/dirmich/maru-bot/releases).
2. Download the `marubot-windows-xxx.zip` or `exe` appropriate for your OS (64-bit or 32-bit).
3. Running the downloaded file will automatically install MaruBot to `~/.marubot/bin` and activate the tray icon.

### 3. 🍎 macOS (GUI)
1. Visit the [Official Release Page](https://github.com/dirmich/maru-bot/releases).
2. Download the `.dmg` file for your CPU architecture (Intel or Apple Silicon).
3. Open the DMG and run `MaruBot.app` to see the icon in your menu bar.

> [!TIP]
> **Note for Windows Users**: Since the binary is not signed with a commercial certificate, you might see a 'Windows protected your PC' warning from SmartScreen. This is expected. To run MaruBot, click **'More info'** and then **'Run anyway'**.

> [!TIP]
> **Note for macOS Users**: For versions without Apple notarization, you may see a "Gatekeeper" warning on first launch. Officially signed versions will run directly; otherwise, go to **'System Settings > Privacy & Security'** and click **'Open Anyway'**, or **'Right-click > Open'** on the app icon.

### 3. Setup API Keys & Default Model
```bash
# Example for OpenAI API Key
marubot config set providers.openai.api_key "YOUR_KEY"

# Select default model (e.g. gpt-4o, gemini-2.5-flash, etc.)
marubot config set agents.defaults.model "gpt-4o"
```
*💡 If the default model fails to connect, MaruBot will automatically fallback to other providers if their API keys are configured!*

### 4. Start Interaction
```bash
marubot agent
```
*(Or use `marubot start` for the Web Dashboard at http://localhost:8080)*

---

## 🧩 Extension Guide (Skills & Tools)
MaruBot allows users to easily expand its capabilities.

### Adding a Tool
To register a simple action or script, just ask the AI directly:
```text
"Write a python script to show system information and register it as a new tool"
```
The AI will autonomously code and save the script with its metadata to the **`extensions` folder**, making it available instantly.

### Adding a Skill
For complex workflows, knowledge bases or prompt templates, you can create a skill:
```text
"Create a new skill for coding templates"
marubot skills show <skill-name>
```
Simply create a directory under `~/.marubot/workspace/skills/<skill-name>` and author a `SKILL.md` file inside it. MaruBot will automatically load and learn it.

---

## 🛠️ Hardware Integration
MaruBot automatically detects your platform and enables supported tools:
- **GPIO**: Control LEDs, buttons, and relays.
- **Camera**: Visual recognition via Libcamera or USB webcams.
- **Sensors**: MPU6050 (IMU), Ultrasonic, GPS, etc.
- **Motors**: Direct control of Servos and DC motors.

---

## 📝 License
MaruBot follows the MIT License. Feel free to modify and distribute.

*Developed & Analyzed by Antigravity AI (2026-03-13)*
