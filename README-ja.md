# MaruBot (マルボット) - 超軽量 AI エージェント (0.6.6) 🦞

**MaruBot** は究極の効率を追求し、わずか **10MBのメモリ (RAM)** 環境でも駆動可能な超軽量フィジカル AI エージェントです。Raspberry Pi から一般的な Linux サーバー、Windows PC までサポートし、自らの能力を拡張する「自己進化」エンジンを備えた最も賢いパーソナルアシスタントです。

---

## ✨ コア機能 (Key Features)

### 1. 🚀 超軽量・高性能 (Ultra-lightweight)
- **10MB RAM:** 最適化された Go バイナリを使用し、組み込みデバイスでも負担なく動作します。
- **シングルバイナリ (Single Binary):** 依存関係のない単一バイナリによる配布で、インストールと管理が非常に簡単です。

### 2. 🌍 マルチプラットフォーム対応 (Multi-Platform)
- **Raspberry Pi:** GPIO、カメラ、各種センサー制御を完全サポート (ARM32/64)。
- **Linux:** Ubuntu、Debian、AWS EC2 など、すべての一般的な Linux 環境をサポート。
- **Windows:** 64bit および 32bit アーキテクチャを公式サポート（バイナリ提供）。
- **macOS:** Intel および Apple Silicon (M1/M2/M3) を公式サポート（DMG 提供）。

### 3. 🧬 自己進化エンジン (Auto-Evolution)
- **`create_tool`**: 新しい原子的な機能を担う Bash/Python スクリプトを自らコーディングし、即座にツールとして登録します。
- **`create_skill`**: 複雑なワークフローやガイドラインを含む上位レベルの「スキル」をフォルダ形式で自動生成します。

### 4. 🧠 スマート記憶システム (RAG Memory)
- **SQLite FTS5:** SQLite と全文検索 (Full-Text Search) を活用し、数万件の対話履歴から関連情報を1秒以内に検索して会話に反映させます。
- **Facts & Preferences:** ユーザーの傾向や重要な事実を永久に記憶します。

### 5. 🛠️ 強力な自動化および制御ツール
- **Cron 予約タスク:** 「明日の午前9時に天気を教えて」のような予約コマンドを実行します。
- **SSH マネージャー:** リモートサーバー接続用の公開鍵の生成と配布を知的に管理します。
- **MAVLink ドローン:** フライトコントローラーと連動し、物理的な飛行制御が可能です。

### 6. 🌐 中央管理およびホームページ (Admin & Landing)
- **公式ホームページ:** 多言語 (KO, EN, JA, ES) 対応およびワンクリックインストールガイドの提供。
- **中央管理 (Admin):** Google SSO と連動し、複数の Marubot インスタンスのステータス（OS、メモリなど）を一目で把握できます。

---

### 🤖 AI モデルとフォールバック (AI Models & Fallback)
MaruBot は複数の AI プロバイダーをサポートしています。`config.json` でプライマリモデルとフォールバックモデルのリストを設定できます。

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
プライマリプロバイダの接続に失敗した場合、MaruBot は自動的にリスト順にフォールバックモデルを試行します。

### 🪟 Windows デプロイメント
Windows ユーザー向けには、[Releases](https://github.com/dirmich/maru-bot/releases) セクションで2種類のディストリビューションを提供しています。
1. **シングルバイナリ (`marubot.exe`)**: すぐに使用できる単一の実行ファイル。
2. **インストール用パッケージ (`marubot-windows-x64.zip`)**: 実行ファイル、デフォルト構成、およびクイックスタートガイドを含むパッケージ。

> [!TIP]
> **Windows ユーザーへの注意**: 商業用証明書で署名されていないバイナリであるため、実行時に SmartScreen の「Windows によって PC が保護されました」という警告が表示される場合があります。**「詳細情報」**をクリックした後、**「実行」**ボタンを選択して続行してください。

> [!TIP]
> **macOS ユーザーへの注意**: Apple の公証 (Notarization) 手続きを経ていないバージョンの場合、初回実行時に「悪意のあるソフトウェアかどうかをシステムで確認できません」という警告が表示されることがあります。公式に署名されたバージョンはすぐに実行可能ですが、そうでない場合は、**「システム設定 > プライバシーとセキュリティ」**から**「このまま開く」**をクリックするか、アプリアイコンで**「右クリック > 開く」**を選択してください。

## 📂 フォルダ構造
- `/config`: MaruBot ハードウェアおよびエージェントのグローバル設定。
- `/skills`: AI エージェントが学習した専門知識およびガイドライン (`SKILL.md`)。
- `/extensions`: `create_tool` で生成された動的スクリプトツール群。
- `/memory`: SQLite 会話データベースおよびファクトストレージ。

---

## 🚀 クイックスタート (Quick Start)

### 1. 🐧 Linux / 🍎 macOS (Terminal)
ターミナルで以下のコマンドを実行することで、すぐにインストールできます (curl が必要です)。
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. 🪟 Windows (GUI/Manual)
Windows ユーザーは、ターミナルコマンドよりも**公式リリースページ**からファイルをダウンロードして実行することをお勧めします。
1. [公式リリースページ (Releases)](https://github.com/dirmich/maru-bot/releases) にアクセスします。
2. 自身の OS (64bit または 32bit) に適した `marubot-windows-xxx.zip` または `exe` ファイルをダウンロードします。
3. ダウンロードしたファイルを実行すると、自動的に `~/.marubot/bin` フォルダにインストールされ、トレイアイコンが有効になります。

### 3. 🍎 macOS (GUI)
1. [公式リリースページ](https://github.com/dirmich/maru-bot/releases) から、自身の CPU (Intel または Apple Silicon) に合わせた `.dmg` ファイルをダウンロードします。
2. DMG ファイルを開き `MaruBot.app` を実行すると、メニューバー（トレイ）にアイコンが表示されます。

### 4. 必須設定 (API キーおよびデフォルトモデルの登録)
```bash
# OpenAI API キーの設定例
marubot config set providers.openai.api_key "YOUR_KEY"

# デフォルトモデルの選択 (例: gpt-4o, gemini-2.5-flash など)
marubot config set agents.defaults.model "gpt-4o"
```
*💡 デフォルトモデルの接続に失敗しても、他のプロバイダの API キーが存在すれば自動的にフォールバックされます！*

### 5. エージェントの開始
```bash
marubot agent
```
*(または `marubot start` を使用して、Web ダッシュボード http://localhost:8080 を利用可能)*

---

## 🧩 拡張ガイド (Skills & Tools)
MaruBot は、ユーザーが簡単に機能を拡張できます。

### ツールの追加 (Adding a Tool)
単一のスクリプト実行など、単純なアクションを登録するには、AI に直接リクエストしてください。
```text
"システム情報を表示するPythonスクリプトを作成して、新しいツールとして登録して"
```
AI が自ら **`extensions` フォルダ** にスクリプトとメタデータを保存し、ランタイム時に即座に利用します。

### スキルの追加 (Adding a Skill)
複雑なワークフローやプロンプトテンプレートの制御が必要な場合は、スキルを作成します。
```text
"新しいコーディングテンプレートのスキルを作って"
marubot skills show <skill-name>
```
物理的に `~/.marubot/workspace/skills/<スキル名>` ディレクトリを作成し、その中に `SKILL.md` を記述するだけで自動的に連携されます。

---

## 🛠️ ハードウェア連動ガイド
MaruBot は、駆動プラットフォームを検出し、利用可能なツールを自動で選別します。
- **GPIO**: LED、ボタン、リレーの制御。
- **Camera**: Libcamera または USB カメラ 통한 視覚認識。
- **Sensors**: MPU6050 (IMU)、超音波センサー、GPS などの連動。
- **Motor**: サーボおよび DC モーターの直接制御。

---

## 📝 ライセンス
MaruBot は MIT ライセンスに従います。ユーザーは自由に変更して配布することができます。

*Developed & Analyzed by Antigravity AI (2026-03-26)*
