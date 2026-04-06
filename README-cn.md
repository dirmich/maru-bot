# MaruBot - 超轻量级 AI 助手 (0.6.5) 🦞

**MaruBot** 是一款追求极致效率的超轻量级物理 AI 助手，能够在仅 **10MB 的内存 (RAM)** 环境中运行。它支持从树莓派 (Raspberry Pi) 到常规的 Linux 服务器乃至 Windows PC。配备了能够扩展自身能力的“自我进化”引擎，它是最智能的个人助手。

---

## ✨ 核心功能 (Key Features)

### 1. 🚀 超轻量与高性能 (Ultra-lightweight)
- **10MB RAM:** 采用优化的 Go 二进制文件，即使在嵌入式设备上也能毫无压力地运行。
- **单文件二进制 (Single Binary):** 无依赖的单个二进制文件分发，使安装和管理变得非常简单。

### 2. 🌍 多平台支持 (Multi-Platform)
- **树莓派 (Raspberry Pi):** 完美支持控制 GPIO、摄像头和各类传感器 (ARM32/64)。
- **Linux:** 支持 Ubuntu、Debian、AWS EC2 等所有常规 Linux 环境。
- **Windows:** 官方支持 64位 和 32位 架构（提供二进制文件）。
- **macOS:** 官方支持 Intel 和 Apple Silicon (M1/M2/M3)（提供 DMG）。

### 3. 🧬 自我进化引擎 (Auto-Evolution)
- **`create_tool`**: 能够自主编写处理新原子级功能的 Bash/Python 脚本，并立即注册为工具。
- **`create_skill`**: 自动以文件夹形式生成包含复杂工作流或指令的高级“技能”。

### 4. 🧠 智能记忆系统 (RAG Memory)
- **SQLite FTS5:** 利用 SQLite 和全文检索 (Full-Text Search)，能在数万条对话记录中 1秒 内检索出相关信息并应用于对话中。
- **Facts & Preferences:** 永久记住用户的偏好、习惯及重要事实。

### 5. 🛠️ 强大的自动化及控制工具
- **Cron 预定任务:** 执行如“明天上午9点告诉我天气”等预定命令。
- **SSH 管理器:** 智能管理用于远程服务器访问的公钥生成与分发。
- **MAVLink 无人机:** 可与飞行控制器联动，实现物理上的飞行控制。

### 6. 🌐 集中管理与主页 (Admin & Landing)
- **官方主页:** 提供多语言 (KO, EN, JA, ES) 支持及一键安装指南。
- **集中管理 (Admin):** 结合 Google SSO，集中查看多个 Marubot 实例的运行状态（操作系统、内存使用等）。

---

### 🤖 AI 模型与后备策略 (AI Models & Fallback)
MaruBot 支持多个 AI 提供商。您可以在 `config.json` 中配置主模型及后备模型列表：

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
如果主提供商连接失败，MaruBot 会自动按顺序尝试列表中的后备模型。

### 🪟 Windows 部署
针对 Windows 用户，我们在 [Releases](https://github.com/dirmich/maru-bot/releases) 页面提供两种版本的发行包：
1. **单一可执行文件 (`marubot.exe`)**: 可直接使用的独立可执行文件。
2. **安装包 (`marubot-windows-x64.zip`)**: 包含可执行文件、默认配置及快速体验指南的压缩包。

> [!TIP]
> **Windows 用户提示**: 由于该二进制文件未经过商业证书签名，在运行时 SmartScreen 可能会提示“Windows 已保护您的电脑”的警告。请点击 **“更多信息 (More info)”**，然后点击 **“仍要运行 (Run anyway)”** 以继续。

> [!TIP]
> **macOS 用户提示**: 对于未经过苹果公证 (Notarization) 的版本，首次运行时可能会弹出“无法验证开发者及恶意软件”的警告。官方签名的版本可直接运行；否则，请进入 **系统设置 > 隐私与安全性** 中点击 **仍要打开**，或在应用程序图标上 **右键 > 打开**。

## 📂 目录结构
- `/config`: MaruBot 硬件及全局智能体配置。
- `/skills`: AI 智能体学习积累的专业知识和操作指南 (`SKILL.md`)。
- `/extensions`: 通过 `create_tool` 生成的动态脚本工具存放目录。
- `/memory`: SQLite 对话数据库和事实存储库。

---

## 🚀 快速开始 (Quick Start)

### 1. 🐧 Linux / 🍎 macOS (Terminal)
在终端中执行以下命令即可立即安装（需要 curl）：
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. 🪟 Windows (GUI/Manual)
对于 Windows 用户，建议从 **官方发布页面** 下载文件运行，而不是使用终端命令：
1. 访问 [官方发布页面 (Releases)](https://github.com/dirmich/maru-bot/releases)。
2. 下载适合您操作系统架构 (64位 或 32位) 的 `marubot-windows-xxx.zip` 压缩包或 `exe` 文件。
3. 双击下载的文件，程序将自动安装至 `~/.marubot/bin` 目录，并激活系统托盘图标。

### 3. 🍎 macOS (GUI)
1. 从 [官方发布页面](https://github.com/dirmich/maru-bot/releases) 下载适配您 CPU 架构（Intel 或 Apple Silicon）的 `.dmg` 安装映像。
2. 打开 DMG 文件并启动 `MaruBot.app`，随即便能在上方菜单栏（系统小托盘）中看到应用图标。

### 4. 必需的设置 (注册 API 密钥及默认模型)
```bash
# OpenAI API 密钥设置示例
marubot config set providers.openai.api_key "YOUR_KEY"

# 设置默认模型 (如: gpt-4o, gemini-2.5-flash 等)
marubot config set agents.defaults.model "gpt-4o"
```
*💡 即使默认模型连接失败，只要配置了其他服务商的 API 密钥，助手就会自动启用备用模型 (Fallback)！*

### 5. 开始智能体对话
```bash
marubot agent
```
*(也可以通过 `marubot start` 启动服务，随后在浏览器访问 http://localhost:8080 使用 Web 控制台)*

---

## 🧩 扩展指南 (Skills & Tools)
MaruBot 允许用户非常轻松地扩展其能力。

### 添加工具 (Adding a Tool)
如需注册单个脚本执行等简单的动作，可直接向 AI 提出请求：
```text
"编写一个用于打印当前系统信息的 Python 脚本，并将其注册为新工具"
```
AI 会自主编写并将脚本与元数据保存至 **`extensions` 目录** 中，并在运行时立即生效利用。

### 添加技能 (Adding a Skill)
如果您的扩展需要控制复杂的工作流或模板提示词，可以创建一个专属技能：
```text
"创建一个提供代码编写模板的新技能"
marubot skills show <skill-name>
```
只需物理创建 `~/.marubot/workspace/skills/<技能名称>` 目录，在其中编写 `SKILL.md`，系统便会自动将其关联并载入。

---

## 🛠️ 硬件联动指南
MaruBot 会自动监测运行的平台，并智能筛选可用的扩展工具：
- **GPIO**: 操控 LED 灯、物理按钮及继电器模块。
- **Camera**: 通过 Libcamera 或 USB 摄像头进行视觉识别。
- **Sensors**: 联动 MPU6050 (IMU 等内置六轴陀螺仪)、超声波测距传感器、GPS 等模块。
- **Motor**: 对舵机 (Servo) 及直流电机 (DC Motor) 进行底层直线控制。

---

## 📝 许可证
MaruBot 遵循 MIT License 进行开源。用户可自由修改及分发。

*Developed & Analyzed by Antigravity AI (2026-03-26)*
