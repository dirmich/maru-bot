# 🤖 MaruBot (丸机器人)

**MaruBot** 是一款基于超轻量级 MaruBot 引擎，专为在树莓派 (Raspberry Pi) 等单板计算机 (SBC) 上直接控制硬件并进行交互而优化的 **"物理 AI 助手 (Physical AI Assistant)"**。

---

## ✨ 核心理念
1. **引擎重用**: 采用高效 Go 二进制文件，保持 RAM 占用低于 10MB。
2. **树莓派优化**: 自动配置 GPIO、摄像头、麦克风和扬声器的权限。
3. **超本地配置**: 无需复杂的 JSON 编辑，通过专属脚本 (`maru-setup.sh`) 以交互方式完成设置。
4. **物理交互**: 内置多种工具，支持 AI 智能体控制舵机、LED 及各类传感器（DHT, PIR 等）。

---

## 📂 目录结构
- `/config`: MaruBot 专属硬件及智能体配置文件
- `maru-setup.sh`: 树莓派初始化及硬件集成自动化脚本
- `/tools`: AI 智能体使用的 GPIO/I2C/SPI 控制工具
- `/bin`: MaruBot 二进制链接或可执行文件

---

## 🚀 快速开始

### 1. 一键安装 (推荐使用 GitHub Gist)
最快速、最简便的安装方式：
```bash
# MaruBot 官方一键安装程序
curl -fsSL https://gist.githubusercontent.com/dirmich/367961d107d6e0f35f1c3156dc55f7d5/raw/install.sh | bash
```

### 2. 手动安装
1. 安装 Go 1.24+ 及必要工具: `sudo apt install -y git make golang libcamera-apps`
2. 克隆仓库: `git clone https://github.com/dirmich/maru-bot.git marubot`
3. 执行设置: `cd marubot && bash maru-setup.sh`

### 3. 运行智能体
```bash
marubot agent
```

### 4. Web 管理后台 (Web Admin Dashboard)
一个可视化的仪表板，用于管理配置、安装技能以及与智能体聊天。
```bash
# 同时启动网关和 Web 管理后台
marubot dashboard
```
在浏览器中访问 `http://localhost:3000` 即可使用管理工具。

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
