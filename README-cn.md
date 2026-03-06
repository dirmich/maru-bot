# MaruBot - Ultra-lightweight AI Assistant (v0.4.11) 🦞

**MaruBot** 是一款基于超轻量级 MaruBot 引擎，专为在树莓派 (Raspberry Pi) 等单板计算机 (SBC) 上直接控制硬件并进行交互而优化的 **"物理 AI 助手 (Physical AI Assistant)"**。

## ✨ 0.4.0 更新
- **新增 Webhook 通道**: 支持外部服务通过 HTTP 请求与 MaruBot 进行实时对话（支持同步响应）。
- **AI 自我进化 (Self-Evolution)**: AI 智能体现在可以自行安装新技能或创建工具，然后通过重启 (`reload`) 自身来自主扩展功能。
- **系统控制工具 (`system_control`)**: 使 AI 能够直接检查 MaruBot 的状态、安装技能并执行进程重启。

---

## ✨ 核心理念
1. **引擎重用**: 采用高效 Go 二进制文件，保持 RAM 占用低于 10MB。
2. **树莓派优化**: 自动配置 GPIO、摄像头、麦克风和扬声器的权限。
3. **超本地配置**: 无需复杂的 JSON 编辑，通过专属脚本 (`maru-setup.sh`) 以交互方式完成设置。
4. **物理交互**: 内置多种工具，支持 AI 智能体控制舵机、LED 及各类传感器（DHT, PIR 等）。

---

## ✨ 0.4.1 更新
- **GPIO 控制改进**: 增加了防止误配置电源 (3.3V, 5V) 和地线引脚的保护逻辑。
- **UI 视觉优化**: 未配置的 GPIO 引脚显示为灰色，增强视觉区分度。
- **Webhook 频道**: 支持通过 HTTP Webhook 与 MaruBot 进行实时对话（支持同步响应）。
- **AI 自我进化 (Self-Evolution)**: 智能体现在可以自主安装技能或创建新工具，并通过重启 (`reload`) 扩展功能。
- **易用性提升**: 在引脚详情视图中增加了“查看全部”按钮，方便返回列表。
- **版本同步**: 将引擎和 Web 管理端的版本同步更新至 0.4.1。

---

## ✨ 0.4.3 更新
- **更新并自动重启**: 改进了 `marubot upgrade`，在更新完成后会自动重启系统。
- **AI 版本控制与更新**: 增加了 AI 智能体检查自身版本并根据用户请求执行升级的工具。
- **Shadcn UI 对话框**: 将 Web Admin 中的所有浏览器原生对话框替换为现代化的 Shadcn UI 对话框，提供一致的 UX。
- **集中式版本管理**: 整合了引擎内部的版本管理逻辑，提高了更新检查和管理的准确性。

---
- **GPIO 配置持久化**: 改进了配置存储，现在可以同时保存 **输入/输出模式 (IN, OUT 等)**，确保在刷新或重启后依然保持设定。
- **引脚选择 UX 优化**: 将原本的操作繁琐的数字输入改为 **下拉选择框 (Select box)**，仅显示可用的 GPIO 引脚，防止冲突并提升操作效率。
- **管理员密码保护**: 优化了 `marubot onboard` 及安装脚本的逻辑，如果输入为空，则自动保留现有密码，避免误操作导致的重置。
- **增强 AI 联动**: 更新了 AI 工具逻辑，使其能够根据用户定义的引脚标签（Label）进行精准识别与控制。

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

最快速启动 MaruBot 的方法。

### 1. 一键安装
在终端执行以下命令，即可一次性完成引擎和 Web 管理后台的下载与安装：

```bash
# Official MaruBot One-Line Installer
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. 必要设置 (注册 API 密钥)
安装完成后，请注册您要使用的 AI 模型的 API 密钥：

```bash
# OpenAI API 密钥设置示例
marubot config set providers.openai.api_key "YOUR_OPENAI_KEY"

# 选择默认模型
marubot config set agents.defaults.model "gpt-4o"
```

### 3. 运行智能体
```bash
# 控制台交互模式
marubot agent

# 或 Web 管理后台仪表板 (http://localhost:3000)
marubot start
```

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
遵循 MIT License。 MaruBot 基于 [picoclaw](https://github.com/sipeed/picoclaw) 开发。

*开发与分析: Antigravity AI (2026)*
