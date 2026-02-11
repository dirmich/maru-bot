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
最快速、最简便的安装方式。请使用在您的 Gist 中点击 **Raw** 按钮获取的 URL：

```bash
# ⚠️ 注意: <GIST_ID> 是您 Gist 页面地址栏最后的字符串。
curl -fsSL https://gist.githubusercontent.com/<USER>/<GIST_ID>/raw/install.sh | bash
```

---

## 📝 开源协议
遵循 MIT License。

MaruBot 基于 [picoclaw](https://github.com/sipeed/picoclaw) 开发，并针对树莓派 (Raspberry Pi) 进行了功能增强。

*开发与分析: Antigravity AI (2026)*
