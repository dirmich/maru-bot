# MaruBot - Ultra-light AI Agent (0.7.2) 🦞

**MaruBot** is an ultra-lightweight, Physical AI agent designed for extreme efficiency, capable of running in an environment with just **10MB of memory (RAM)**. It supports ranging from Raspberry Pi to generic Linux servers and Windows PCs. Equipped with a 'Self-Evolution' engine that expands its own capabilities, it is the smartest personal assistant.

---

## ✨ Key Features

### 1. 🚀 Ultra-lightweight & High Performance
- **10MB RAM:** Operates seamlessly even on embedded devices using an optimized Go binary.
- **Single Binary:** Extremely simple installation and management with dependency-free single-binary deployment.

### 2. 🌍 Multi-Platform Support
- **Raspberry Pi:** Full support for controlling GPIO, cameras, and various sensors (ARM32/64).
- **Linux:** Supports all standard Linux environments including Ubuntu, Debian, and AWS EC2.
- **Windows:** Officially supports 64-bit and 32-bit architectures (binaries provided).
- **macOS:** Officially supports Intel and Apple Silicon (M1/M2/M3) (DMG provided).

### 3. 🧬 Auto-Evolution Engine
- **`create_tool`**: Autonomously codes Bash/Python scripts for new atomic functions and registers them as tools instantly.
- **`create_skill`**: Automatically generates high-level folder-based 'skills' containing complex workflows or instructions.

### 4. 🧠 Smart Memory System (RAG Memory)
- **SQLite FTS5:** Retrieves relevant information within 1 second from tens of thousands of conversations by utilizing SQLite and Full-Text Search.
- **Facts & Preferences:** Permanently remembers user traits, preferences, and important facts.

### 5. 🛠️ Powerful Automation & Control Tools
- **Cron Scheduled Tasks:** Executes scheduled commands like "Tell me the weather at 9 AM tomorrow."
- **SSH Manager:** Intelligently manages the generation and deployment of public keys for remote server access.
- **MAVLink Drone:** Interlocks with flight controllers for physical drone control.

### 6. 🌐 Centralized Management & Landing Page
- **Official Landing Page:** Multilingual support (KO, EN, JA, ES) and one-click installation guide provided.
- **Admin Dashboard:** Integrates with Google SSO to view the status (OS, memory, etc.) of multiple Marubot instances at a glance.

---

### 🤖 AI Models & Fallback
MaruBot supports multiple AI providers. You can configure a primary model and a list of fallback models in your `config.json`:

```json
{
  "agents": {
    "defaults": {
      "model": "glm-4.7",
      "fallback_models": ["gpt-4o", "claude-3-5-sonnet", "gemini-2.0-flash"]
    }
  }
}
```
If the primary provider fails, MaruBot will automatically try the fallback models in order.

### 🪟 Windows Deployment
For Windows users, we provide two types of distributions in the [Releases](https://github.com/dirmich/maru-bot/releases) section:
1. **Single Binary (`marubot.exe`)**: A standalone executable for quick use.
2. **Installable Package (`marubot-windows-x64.zip`)**: Includes the executable, default configuration, and a quick-start guide.

> [!TIP]
> **Note for Windows users**: Because the binary is not signed with a commercial certificate, you may see a SmartScreen 'Windows protected your PC' warning when running it. Please click **'More info'** and then select the **'Run anyway'** button to proceed.

> [!TIP]
> **Note for macOS users**: For versions that have not undergone Apple's notarization process, a warning stating 'cannot be checked for malicious software' may appear upon first launch. Officially signed versions can be run directly, but if not, click **'Open Anyway'** in **'System Settings > Privacy & Security'**, or select **'Right-click > Open'** on the app icon.

## 📂 Project Structure
- `/config`: MaruBot hardware and global agent configurations.
- `/skills`: Specialized knowledge and guidelines learned by the AI agent (`SKILL.md`).
- `/extensions`: Dynamic script tools created via `create_tool`.
- `/memory`: SQLite conversation database and fact storage.

---

## 🚀 Quick Start

### 1. 🐧 Linux / 🍎 macOS (Terminal)
You can install instantly by running the command below in your terminal (requires curl):
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. 🪟 Windows (GUI/Manual)
For Windows users, we recommend downloading the file from the **Official Releases Page** instead of using terminal commands:
1. Visit the [Official Releases Page](https://github.com/dirmich/maru-bot/releases).
2. Download the `marubot-windows-xxx.zip` or `exe` file matching your OS (64-bit or 32-bit).
3. If you run the downloaded file, it will be automatically installed in the `~/.marubot/bin` folder, and the tray icon will be activated.

### 3. 🍎 macOS (GUI)
1. Download the `.dmg` file matching your CPU (Intel or Apple Silicon) from the [Official Releases Page](https://github.com/dirmich/maru-bot/releases).
2. Open the DMG file and run `MaruBot.app` to see the icon in the menu bar (tray).

### 4. Required Settings (API Key & Default Model Setup)
```bash
# Example of setting OpenAI API Key
marubot config set providers.openai.api_key "YOUR_KEY"

# Select default model (e.g., gpt-4o, gemini-2.5-flash, etc.)
marubot config set agents.defaults.model "gpt-4o"
```
*💡 Even if connection to the primary model fails, it will automatically fallback if API keys for other providers exist!*

### 5. Start Agent Conversation
```bash
marubot agent
```
*(Alternatively, you can use the web dashboard at http://localhost:8080 by running `marubot start`)*

---

## 🧩 Extension Guide (Skills & Tools)
MaruBot allows users to easily expand its capabilities.

### Adding a Tool
To register a simple action such as running a single script, ask the AI directly.
```text
"Write a python script that shows system information and register it as a new tool"
```
The AI will autonomously save the script and metadata in the **`extensions` folder** and immediately utilize it at runtime.

### Adding a Skill
Create a skill if you need to control complex workflows or prompt templates.
```text
"Create a new coding template skill"
marubot skills show <skill-name>
```
Simply create the `~/.marubot/workspace/skills/<skill-name>` directory physically and write `SKILL.md` inside it to have it automatically linked.

---

## 🛠️ Hardware Integration Guide
MaruBot detects the running platform and automatically selects available tools.
- **GPIO**: LED, button, and relay control.
- **Camera**: Visual recognition through Libcamera or USB cameras.
- **Sensors**: Interlock with MPU6050 (IMU), ultrasonic sensors, GPS, etc.
- **Motor**: Direct control of servo and DC motors.

---

## 📝 License
MaruBot follows the MIT License. Users are free to modify and distribute.

*Developed & Analyzed by Antigravity AI (2026-03-24)*
