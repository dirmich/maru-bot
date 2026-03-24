# MaruBot - 超轻量级 AI 助手 (0.5.8) 🦞
**MaruBot** 是一款基于超轻量级 MaruBot 引擎，专为在树莓派 (Raspberry Pi) 等单板计算机 (SBC) 上直接控制硬件并进行交互而优化的 **"物理 AI 助手 (Physical AI Assistant)"**。

## ✨ 核心理念
1. **引擎重用**: 采用高效 Go 二进制文件，保持 RAM 占用低于 10MB。
2. **树莓派优化**: 自动配置 GPIO、摄像头、麦克风和扬声器的权限。
3. **超本地配置**: 无需复杂的 JSON 编辑，通过专属脚本 (`maru-setup.sh`) 以交互方式完成设置。
4. **物理交互**: 内置多种工具，支持 AI 智能体控制舵机、LED 及各类传感器（DHT, PIR 等）。
5. **中央管理与主页**: 提供支持多语言的主页，以及通过 Google SSO 登录的多个实例的中央管理控制面板。

---

## 📂 目录结构
- `/config`: MaruBot 专属硬件及智能体配置文件
- `maru-setup.sh`: 树莓派初始化及硬件集成自动化脚本
- `/tools`: AI 智能体使用的 GPIO/I2C/SPI 控制工具
- `/bin`: MaruBot 二进制链接或可执行文件

---

## 📋 事前准备 (Prerequisites)

在开始之前，请确保您已准备好以下内容：
- **Hardware**: 树莓派 (完美支持 ARM64/32), 电源适配器, SD 卡
- **OS**: 树莓派 OS (建议 Bullseye 或更高版本)
- **API Key**: OpenAI, Gemini 等 LLM 服务的 API 密钥

---

## 🚀 快速开始 (Quick Start)

最快开始 MaruBot 的方式。

### 1. 🐧 Linux / 🍎 macOS (Terminal)
在终端执行以下命令即可立即安装 (需要 curl):
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. 🪟 Windows (GUI/Manual)
Windows 用户建议从 **官方发布页面** 下载文件，而不是使用终端命令:
1. 访问 [官方发布页面 (Releases)](https://github.com/dirmich/maru-bot/releases)。
2. 选择适合您系统的 `marubot-windows-xxx.zip` 或 `exe` 文件 (64位或32位)。
3. 运行下载的文件，它会自动安装到 `~/.marubot/bin` 并激活任务栏图标。

---

## 🛠️ 详细安装及硬件联动

如果一键安装程序无法正常工作，或者您更倾向于手动设置：

1. **安装基础工具**: `sudo apt install -y git make golang libcamera-apps`
2. **克隆仓库**: `git clone https://github.com/dirmich/maru-bot.git marubot`
3. **硬件初始化**: `cd marubot && bash maru-setup.sh`
   - 此脚本将检查 GPIO 权限并激活摄像头/音频接口。

---

## ⚙️ 配置 (Configuration)

1. **使用命令行工具 (推荐)**:
   ```bash
   # 设置 OpenAI API 密钥
   marubot config set providers.openai.api_key "YOUR_KEY"
   
   # 修改默认模型
   marubot config set agents.defaults.model "gpt-4o"
   ```

2. **手动修改配置文件**:
   ```bash
   nano ~/.marubot/config.json
   ```
   在 `providers` 部分填写相应服务的 `api_key`。

---

## 📝 开源协议
遵循 MIT License。

*开发与分析: Antigravity AI (2026-03-24)*
