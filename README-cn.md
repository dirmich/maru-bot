# 🦞 MaruBot - 超轻量级 AI 助手 (v0.9.5)

[![Official Website](https://img.shields.io/badge/Official-Website-blue?style=for-the-badge&logo=googlechrome&logoColor=white)](https://dirmich.github.io/maru-bot/)
[![GitHub Release](https://img.shields.io/badge/Release-v0.9.5-emerald?style=for-the-badge)](https://github.com/dirmich/maru-bot/releases)

> **"当智能触碰现实 — MaruBot"**
> 
> MaruBot 不仅仅是一个聊天机器人。它是一个基于边缘计算的 AI 助手，能够智能地控制物理硬件 (GPIO) 并执行自主技能。
> 
> **[👉 在官方落地页了解更多 (安装指南)](https://dirmich.github.io/maru-bot/)**

---

## 🚀 核心功能 (Key Features)

### 1. ⚡ 超轻量与高性能 (Ultra-lightweight)
- **10MB RAM:** 采用优化的 Go 二进制文件，即使在嵌入式设备上也能轻压运行。
- **单二进制文件 (Single Binary):** 无依赖的单个二进制文件分发，使安装和管理变得非常简单。

### 2. 🌐 多平台支持 (Multi-Platform)
- **树莓派 (Raspberry Pi):** 完美支持控制 GPIO、摄像头及各类传感器 (ARM32/64)。
- **Linux:** 支持所有标准的 Linux 环境（Ubuntu、Debian、AWS EC2 等）。
- **Windows:** 官方支持 64 位及 32 位架构。
- **macOS:** 官方支持 Intel 及 Apple Silicon (M1/M2/M3)。

### 3. 🧠 硬件原生控制 (Hardware Native)
- **GPIO 控制面板:** 实时监测和控制树莓派的引脚状态。
- **技能引擎 (Skill Engine):** 自动执行连接系统命令与硬件操作的自主技能。

---

## 🎬 快速开始 (Quick Start)

在终端运行以下命令即可立即安装 (需要 curl)：
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

请访问 [官方落地页](https://dirmich.github.io/maru-bot/) 获取更多信息。

*Handcrafted by Highmaru (2026)*