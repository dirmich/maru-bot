<<<<<<< HEAD
# MaruBot - Ultra-lightweight AI Assistant (0.4.86) 🦞
=======
# MaruBot - 超軽量 AI アシスタント (0.4.87) 🦞
>>>>>>> 42feaf4 (슬랙 연동 로깅 개선 및 0.4.87 릴리즈)

**MaruBot** は、超軽量の MaruBot エンジンをベースに、Raspberry Pi などの SBC（シングルボードコンピュータ）でハードウェアを直接制御し、対話するために最適化された **"Physical AI Assistant"** です。

## ✨ 核心コンセプト
1. **エンジンの再利用**: 高効率な Go バイナリを使用し、RAM 占有率を 10MB 以下に維持します.
2. **Raspberry Pi 最適化**: GPIO、カメラ、マイク、スピーカーの権限設定을 自動化します。
3. **ハイパーローカル設定**: 複雑な JSON 編集の代わりに、専用スクリプト (`maru-setup.sh`) を通じて対話形式で設定を完了します。
4. **物理的な相互作用**: サーボモーター、LED、各種センサー（DHT、PIR など）を AI エージェントが制御できるツールが事前に含まれています。
5. **中央管理とホームページ**: マルチリンガル対応のホームページと、Google SSO を使用した複数インスタンスの中央管理ダッシュボードを提供します。

---

## 📂 フォルダ構造
- `/config`: MaruBot 専用のハードウェアおよびエージェント設定ファイル
- `maru-setup.sh`: Raspberry Pi の初期化 및 ハードウェア連携自動化スクリプト
- `/tools`: AI エージェントが使用する GPIO/I2C/SPI 制御ユーティリティ
- `/bin`: MaruBot バイナリリンク 또는 実行ファイルの保管

---

## 📋 事前準備 (Prerequisites)

開始する前に、以下の準備ができているか確認してください：
- **Hardware**: Raspberry Pi (ARM64/32 完全対応), 電源アダプター, SDカード
- **OS**: Raspberry Pi OS (Bullseye 以上推奨)
- **API Key**: OpenAI, Gemini など使用する LLM サービスの API キー

---

## 🚀 クイックスタート (Quick Start)

最も素早く MaruBot を開始する方法です。

### 1. 🐧 Linux / 🍎 macOS (Terminal)
터미널で以下のコマンドを実行して、すぐにインストールできます (curl が必要です):
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. 🪟 Windows (GUI/Manual)
Windows ユーザーは、ターミ널コマンドよりも **公式リリースページ** からファイルをダウンロードして実行することをお勧めします:
1. [公式リリースページ (Releases)](https://github.com/dirmich/maru-bot/releases) にアクセスします。
2. お使いの OS (64bit または 32bit) に適した `marubot-windows-xxx.zip` または `exe` ファイルをダウンロードします.
3. ダウンロードしたファイルを実行すると、自動的に `~/.marubot/bin` フォルダにインストールされ、トレイアイコンが有効になります。

### 3. 必須設定 (API キーの登録)
インストール完了後、使用する AI モデルの API キーを登録します：

```bash
# OpenAI API キーの設定例
marubot config set providers.openai.api_key "YOUR_OPENAI_KEY"

# デフォルトモデルの選択
marubot config set agents.defaults.model "gpt-4o"
```

### 4. エージェントの実行
```bash
# コンソール対話モード
marubot agent

# または Web 管理者ダッシュボード (http://localhost:3000)
marubot start
```

---

## 🛠️ 詳細インストールおよびハードウェア連携

ワンクリックインストーラーが動作しない場合や、手動設定を希望する場合：

1. **必須ツールのインストール**: `sudo apt install -y git make golang libcamera-apps`
2. **リポジトリのクローン**: `git clone https://github.com/dirmich/maru-bot.git marubot`
3. **ハードウェア初期化**: `cd marubot && bash maru-setup.sh`
   - このスクリプトは GPIO 権限、カメラ/オーディオインターフェースの有効化を確認します。

---

## ⚙️ 設定 (Configuration)

1. **CLIを使用する (推奨)**:
   ```bash
   # OpenAI APIキーを設定
   marubot config set providers.openai.api_key "YOUR_KEY"
   
   # デフォルトモデルを変更
   marubot config set agents.defaults.model "gpt-4o"
   ```

2. **設定ファイルを直接編集**:
   ```bash
   nano ~/.marubot/config.json
   ```
   `providers` セクションで使用するサービスの `api_key` を入力。

---

## 📝 ライセンス
MIT License。

*開発と分析: Antigravity AI (2026-03-20)*
