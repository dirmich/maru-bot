# 🤖 MaruBot (マルボット)

**MaruBot** は、超軽量の MaruBot エンジンをベースに、Raspberry Pi などの SBC（シングルボードコンピュータ）でハードウェアを直接制御し、対話するために最適化された **"Physical AI Assistant"** です。

---

## ✨ コアコンセプト
1. **エンジンの再利用**: 高効率な Go バイナリを使用し、RAM 占有率を 10MB 以下に維持します。
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
最も速く簡単なインストール方法です。自身の Gist で **Raw** ボタンを押して取得した URL を使用してください：

```bash
# MaruBot 公式ワンクリックインストーラー
curl -fsSL https://gist.githubusercontent.com/dirmich/367961d107d6e0f35f1c3156dc55f7d5/raw/install.sh | bash
```

---

## 📝 ライセンス
MIT License。

MaruBot は [picoclaw](https://github.com/sipeed/picoclaw) をベースにしており、Raspberry Pi 用に最適化された機能が追加されています。

*開発と分析: Antigravity AI (2026)*
