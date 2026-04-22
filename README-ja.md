# 🦞 MaruBot - 超軽量 AI エージェント (v0.9.5)

[![Official Website](https://img.shields.io/badge/Official-Website-blue?style=for-the-badge&logo=googlechrome&logoColor=white)](https://dirmich.github.io/maru-bot/)
[![GitHub Release](https://img.shields.io/badge/Release-v0.9.5-emerald?style=for-the-badge)](https://github.com/dirmich/maru-bot/releases)

> **"知性と実体が交차する場所 — MaruBot"**
> 
> MaruBot は単なるチャットボットではありません。Edge コンピューティングに基づき、物理的なハードウェア (GPIO) を知的に制御し、自律的なスキルを実行する AI エージェントです。
> 
> **[👉 公式ランディングページで詳細を確認 (インストールガイド)](https://dirmich.github.io/maru-bot/)**

---

## 🚀 主な機能 (Key Features)

### 1. ⚡ 超軽量・高性能 (Ultra-lightweight)
- **10MB RAM:** 最適化された Go バイナリを使用し、組み込みデバイスでも低負荷で動作します。
- **シングルバイナリ (Single Binary):** 依存関係のない単一のバイナリ配布により、インストールと管理が非常に簡単です。

### 2. 🌐 マルチプラットフォーム対応 (Multi-Platform)
- **Raspberry Pi:** GPIO、カメラ、各種センサー制御を完全にサポート (ARM32/64)。
- **Linux:** すべての標準的な Linux 環境（Ubuntu、Debian、AWS EC2 など）をサポート。
- **Windows:** 64bit および 32bit アーキテクチャを公式サポート。
- **macOS:** Intel および Apple Silicon (M1/M2/M3) を公式サポート。

### 3. 🧠 ハードウェアネイティブ制御 (Hardware Native)
- **GPIO ダッシュボード:** Raspberry Pi のピン状態をリアルタイムで監視・制御。
- **スキルエンジン (Skill Engine):** システムコマンドとハードウェア動作を連携させる自律的なスキル実行。

---

## 🎬 クイックスタート (Quick Start)

ターミナルで以下のコマンドを実行して即座にインストールできます (curl が必要):
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

詳細は [公式ランディングページ](https://dirmich.github.io/maru-bot/) を参照してください。

*Handcrafted by Highmaru (2026)*