# 🤖 MaruBot (マルボット)

**MaruBot** は、超軽量の MaruBot エンジンをベースに、Raspberry Pi などの SBC（シングルボードコンピュータ）でハードウェアを直接制御し、対話するために最適化された **"Physical AI Assistant"** です。

---

## ✨ コアコンセプト
1. **エンジンの再利用**: 高効率な Go バイナリを使用し、RAM 占有率를 10MB 以下に維持します。
2. **Raspberry Pi 最適化**: GPIO、カメラ、マイク、スピーカーの権限設定を自動化します。
3. **ハイパーローカル設定**: 複雑な JSON 編集の代わりに、専用スクリプト (`maru-setup.sh`) を通じて対話形式で設定を完了します。
4. **物理的な相互作用**: サーボモーター、LED、各種センサー（DHT、PIR など）を AI エージェントが制御できるツールが事前に含まれています。

---

## 📂 フォルダ構造
- `/config`: MaruBot 専用のハードウェアおよびエージェント設定ファイル
- `maru-setup.sh`: Raspberry Pi の初期化およびハードウェア連携自動化スクリプト
- `/tools`: AI エージェントが使用する GPIO/I2C/SPI 制御ユーティリティ
- `/bin`: MaruBot バイナリリンクまたは実行ファイルの保管

---

## 🚀 クイックスタート

### 1. ワンクリックインストール (GitHub Gist 推奨)
最も速く簡単なインストール方法です：
```bash
# MaruBot 公式ワンクリックインストーラー
curl -fsSL https://gist.githubusercontent.com/dirmich/367961d107d6e0f35f1c3156dc55f7d5/raw/install.sh | bash
```

### 2. 手動インストール
1. Go 1.24+ および必須ツールのインストール: `sudo apt install -y git make golang libcamera-apps`
2. リポジトリのクローン: `git clone https://github.com/dirmich/maru-bot.git marubot`
3. セットアップの実行: `cd marubot && bash maru-setup.sh`

### 3. エージェントの実行
```bash
marubot agent
```

### 4. Web Admin ダッシュボード
設定の管理、スキルのインストール、エージェントとのチャットを視覚的に行えるウェブダッシュボードです。
```bash
# ゲートウェイとウェブダッシュボードを同時に起動
marubot dashboard
```
ブラウザから `http://localhost:3000` にアクセスして管理ツールを使用できます。

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
MIT License。 MaruBot は [picoclaw](https://github.com/sipeed/picoclaw) をベースにしています。

*開発と分析: Antigravity AI (2026)*
