import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type Language = 'en' | 'ko' | 'ja';

interface TranslationStore {
    language: Language;
    setLanguage: (lang: Language) => void;
}

export const useLanguageStore = create<TranslationStore>()(
    persist(
        (set) => ({
            language: 'ko', // Default to Korean
            setLanguage: (lang) => set({ language: lang }),
        }),
        {
            name: 'marubot-language',
        }
    )
);

export type TranslationKey = keyof typeof translations.en;

export const translations = {
    en: {
        // Sidebar
        dashboard: "Dashboard",
        chat: "AI Assistant",
        gpio: "GPIO Control",
        skills: "Skills & Tools",
        settings: "Settings",
        logs: "System Logs",
        logs_desc: "View background activity and diagnostic records.",

        // Chat Page
        chat_title: "AI Assistant",
        chat_desc: "Talk with the agent in real-time.",
        chat_live: "Live Chat",
        chat_clear_confirm_title: "Clear Chat History",
        chat_clear_confirm_desc: "All chat history will be deleted. Do you want to continue?",
        chat_input_placeholder: "Type a message...",
        chat_empty_msg: "Type a message to start conversation.",
        chat_thinking: "Thinking...",
        chat_send_error: "Failed to send message. (Check if offline)",
        chat_clear_success: "Chat history cleared.",
        chat_welcome: "Hello! I am MaruBot AI Assistant. How can I help you?",

        // GPIO Page
        gpio_title: "GPIO Control & Settings",
        gpio_desc: "View Raspberry Pi pin map and configure hardware interfaces.",
        gpio_schematic: "Pin Map Schematic",
        gpio_schematic_desc: "Click on pins for details.",
        gpio_configured_devices: "Configured Devices",
        gpio_configured_desc: "List of active GPIO pins.",
        gpio_add: "Add",
        gpio_pin: "Pin",
        gpio_mode: "Mode",
        gpio_label: "Label",
        gpio_save: "Save Settings",
        gpio_save_success: "GPIO settings saved.",
        gpio_no_pins: "No GPIO pins configured.",
        gpio_view_all: "View All",
        gpio_cannot_configure: "Power and Ground pins cannot be configured.",
        gpio_add_confirm_title: "Add GPIO Configuration",
        gpio_add_confirm_desc: "Do you want to add a configuration for pin",
        gpio_add_confirm_suffix: "?",

        // Skills Page
        skills_title: "Skills & Tool Box",
        skills_desc: "Manage tools that extend the agent's capabilities.",
        skills_empty: "Skills list empty",
        skills_uninstall: "Uninstall",
        refresh: "Refresh",
        skills_cli_output: "CLI Output",
        skills_install: "Install",
        skills_install_placeholder: "GitHub user/repo",
        skills_install_success: "Skill installed successfully.",
        skills_remove_success: "Skill removed.",
        skills_installing: "Installing...",
        skills_confirm_title_install: "Install Skill",
        skills_confirm_title_remove: "Remove Skill",
        skills_confirm_desc_install: "Do you want to install [{skill}]?",
        skills_confirm_desc_remove: "Do you want to remove [{skill}]?",

        // Settings Page
        settings_title: "Configuration",
        settings_desc: "Manage engine and AI service settings.",
        settings_agent_title: "Main Agent",
        settings_agent_desc: "Set default model and workspace.",
        settings_model: "Model",
        settings_workspace: "Workspace",
        settings_providers_title: "API Providers",
        settings_providers_desc: "Enter API keys for AI services.",
        settings_api_key: "API KEY",
        settings_api_base: "API BASE (Optional)",
        settings_security_title: "Security & Auth",
        settings_security_desc: "Manage admin permissions.",
        settings_admin_account: "Admin Account",
        settings_change_password: "Change Password",
        settings_reset: "Reset Settings",
        settings_save_btn: "Save Config",
        settings_save_confirm_title: "Save Config",
        settings_save_confirm_desc: "Do you want to save changes?",
        settings_reset_confirm_title: "Reset Config",
        settings_reset_confirm_desc: "Do you want to reset all settings?",
        settings_save_success: "Settings saved.",
        settings_reset_not_impl: "Reset function is not yet implemented.",

        settings_channels_title: "Social Channels",
        settings_channels_desc: "Configure bot connectivity for Discord, Telegram, etc.",
        settings_channel_enabled: "Enabled",
        settings_channel_token: "Bot Token",
        settings_channel_allow_from: "Allowed Users/IDs (Comma separated)",

        // Common
        loading: "Loading...",
        save: "Save",
        cancel: "Cancel",
        confirm: "Confirm",
        delete: "Delete",
        status_ok: "SYSTEM READY",

        // Login
        login_title: "Enter your admin password to access the dashboard",
        admin_password: "Admin Password",
        login_btn: "Login",
        logging_in: "Logging in...",
        login_success: "Login successful",
        login_failed: "Invalid password",
        conn_error: "Connection error",

        // Upgrade
        upgrade_available: "New version available",
        upgrade_button: "Upgrade Now",
        upgrading: "Upgrading... System will restart.",
        upgrade_confirm: "Do you want to upgrade to the latest version?",
        
        // Setup Notice
        setup_notice_title: "Initial Configuration Required",
        setup_notice_desc: "MaruBot requires at least one AI model and one social channel to be configured to function properly.",
        setup_notice_ai: "AI Model: {status}",
        setup_notice_channel: "Social Channel: {status}",
        setup_notice_configured: "Configured",
        setup_notice_not_configured: "Not Configured",
        setup_notice_go_settings: "Go to Settings",
    },
    ko: {
        // Sidebar
        dashboard: "대시보드",
        chat: "AI 어시스턴트",
        gpio: "GPIO 제어",
        skills: "스킬 & 툴 박스",
        settings: "환경 설정",
        logs: "시스템 로그",
        logs_desc: "백그라운드 활동 및 진단 기록을 확인합니다.",

        // Chat Page
        chat_title: "AI 어시스턴트",
        chat_desc: "에이전트와 실시간으로 대화하세요.",
        chat_live: "실시간 대화",
        chat_clear_confirm_title: "채팅 내역 삭제",
        chat_clear_confirm_desc: "모든 채팅 내역이 삭제됩니다. 계속하시겠습니까?",
        chat_input_placeholder: "메시지를 입력하세요...",
        chat_empty_msg: "메시지를 입력하여 대화를 시작하세요.",
        chat_thinking: "생각 중...",
        chat_send_error: "메시지 전송에 실패했습니다. (오프라인 모드일 수 있습니다)",
        chat_clear_success: "채팅 내역이 초기화되었습니다.",
        chat_welcome: "안녕하세요! MaruBot AI 어시스턴트입니다. 무엇을 도와드릴까요?",

        // GPIO Page
        gpio_title: "GPIO 제어 및 설정",
        gpio_desc: "Raspberry Pi의 핀 맵을 시각적으로 확인하고 하드웨어 인터페이스를 설정합니다.",
        gpio_schematic: "핀 맵 스케매틱",
        gpio_schematic_desc: "핀 번호를 클릭하여 상세 정보를 확인하세요.",
        gpio_configured_devices: "설정된 장치",
        gpio_configured_desc: "활성화된 GPIO 핀 목록입니다.",
        gpio_add: "추가",
        gpio_pin: "Pin",
        gpio_mode: "Mode",
        gpio_label: "Label",
        gpio_save: "설정 저장",
        gpio_save_success: "GPIO 설정이 저장되었습니다.",
        gpio_no_pins: "설정된 GPIO 핀이 없습니다.",
        gpio_view_all: "전체 보기",
        gpio_cannot_configure: "전원 및 그라운드 핀은 설정할 수 없습니다.",
        gpio_add_confirm_title: "GPIO 설정 추가",
        gpio_add_confirm_desc: "해당 핀에 대한 설정을 추가하시겠습니까? (핀 번호: ",
        gpio_add_confirm_suffix: ")",

        // Skills Page
        skills_title: "스킬 & 툴 박스",
        skills_desc: "에이전트의 기능을 확장하는 도구를 관리합니다.",
        skills_empty: "설치된 스킬이 없습니다.",
        skills_uninstall: "삭제",
        refresh: "새로고침",
        skills_cli_output: "CLI 출력",
        skills_install: "설치",
        skills_install_placeholder: "GitHub user/repo",
        skills_install_success: "스킬 설치 완료",
        skills_remove_success: "스킬 삭제 완료",
        skills_installing: "설치 중...",
        skills_confirm_title_install: "스킬 설치",
        skills_confirm_title_remove: "스킬 삭제",
        skills_confirm_desc_install: "[{skill}]을(를) 설치하시겠습니까?",
        skills_confirm_desc_remove: "[{skill}]을(를) 삭제하시겠습니까?",

        // Settings Page
        settings_title: "환경 설정",
        settings_desc: "엔진 및 AI 서비스 설정을 관리합니다.",
        settings_agent_title: "메인 에이전트",
        settings_agent_desc: "기본 동작 모델과 작업 디렉토리를 설정합니다.",
        settings_model: "사용 모델",
        settings_workspace: "워크스페이스",
        settings_providers_title: "API 제공자",
        settings_providers_desc: "연동할 AI 모델 서비스의 인증 키를 입력하세요.",
        settings_api_key: "API KEY",
        settings_api_base: "API BASE (선택 사항)",
        settings_security_title: "보안 및 인증",
        settings_security_desc: "관리자 권한을 설정합니다.",
        settings_admin_account: "관리자 계정",
        settings_change_password: "비밀번호 변경",
        settings_reset: "설정 초기화",
        settings_save_btn: "설정 저장",
        settings_save_confirm_title: "설정 저장",
        settings_save_confirm_desc: "변경사항을 저장하시겠습니까?",
        settings_reset_confirm_title: "설정 리셋",
        settings_reset_confirm_desc: "모든 설정을 초기화하시겠습니까?",
        settings_save_success: "설정이 저장되었습니다.",
        settings_reset_not_impl: "초기화 기능은 아직 구현되지 않았습니다.",

        settings_channels_title: "메신저 채널 연동",
        settings_channels_desc: "디스코드, 텔레그램 등 외부 메신저 연동 설정을 관리합니다.",
        settings_channel_enabled: "활성화함",
        settings_channel_token: "봇 토큰 (Bot Token)",
        settings_channel_allow_from: "허용할 사용자/채팅방 ID (쉼표로 구분)",

        // Common
        loading: "로딩 중...",
        save: "저장",
        cancel: "취소",
        confirm: "확인",
        delete: "삭제",
        status_ok: "시스템 준비 완료",

        // Login
        login_title: "대시보드에 접근하려면 관리자 암호를 입력하세요",
        admin_password: "관리자 암호",
        login_btn: "로그인",
        logging_in: "로그인 중...",
        login_success: "로그인 성공",
        login_failed: "잘못된 암호입니다",
        conn_error: "연결 오류",

        // Upgrade
        upgrade_available: "새 버전 사용 가능",
        upgrade_button: "지금 업그레이드",
        upgrading: "업그레이드 중... 시스템이 재시작됩니다.",
        upgrade_confirm: "최신 버전으로 업그레이드하시겠습니까?",

        // Setup Notice
        setup_notice_title: "초기 설정 필요",
        setup_notice_desc: "MaruBot이 정상 작동하려면 최소 한 개 이상의 AI 모델과 소셜 채널이 설정되어야 합니다.",
        setup_notice_ai: "AI 모델 설정: {status}",
        setup_notice_channel: "소셜 채널 설정: {status}",
        setup_notice_configured: "설정됨",
        setup_notice_not_configured: "설정 안 됨",
        setup_notice_go_settings: "설정하러 가기",
    },
    ja: {
        // Sidebar
        dashboard: "ダッシュボード",
        chat: "AIアシスタント",
        gpio: "GPIO制御",
        skills: "スキル＆ツール",
        settings: "構成設定",
        logs: "システムログ",
        logs_desc: "バックグラウンド活動と診断記録を確認します。",

        // Chat Page
        chat_title: "AIアシスタント",
        chat_desc: "エージェントとリアルタイムで対話します。",
        chat_live: "ライブチャット",
        chat_clear_confirm_title: "チャット履歴の削除",
        chat_clear_confirm_desc: "すべてのチャット履歴が削除されます。続行しますか？",
        chat_input_placeholder: "メッセージを入力...",
        chat_empty_msg: "メッセージを入力して会話を開始します。",
        chat_thinking: "考え中...",
        chat_send_error: "メッセージの送信に失敗しました。（オフラインの可能性があります）",
        chat_clear_success: "チャット履歴がクリアされました。",
        chat_welcome: "こんにちは！MaruBot AIアシスタントです。何かお手伝いしましょうか？",

        // GPIO Page
        gpio_title: "GPIO制御と設定",
        gpio_desc: "Raspberry Piのピンマップを視覚的に確認し、ハードウェアインターフェースを設定します。",
        gpio_schematic: "ピンマップ回路図",
        gpio_schematic_desc: "ピンをクリックして詳細を確認してください。",
        gpio_configured_devices: "設定済みデバイス",
        gpio_configured_desc: "アクティブなGPIOピンのリストです。",
        gpio_add: "追加",
        gpio_pin: "ピン",
        gpio_mode: "モード",
        gpio_label: "ラベル",
        gpio_save: "設定を保存",
        gpio_save_success: "GPIO設定が保存されました。",
        gpio_no_pins: "設定されたGPIOピンはありません。",
        gpio_view_all: "すべて表示",
        gpio_cannot_configure: "電源とグランドピンは設定できません。",
        gpio_add_confirm_title: "GPIO設定の追加",
        gpio_add_confirm_desc: "このピンの設定を追加しますか？ (ピン: ",
        gpio_add_confirm_suffix: ")",

        // Skills Page
        skills_title: "スキル＆ツールボックス",
        skills_desc: "エージェントの機能を拡張するツールを管理します.",
        skills_empty: "スキルリストは空です",
        skills_uninstall: "アンインストール",
        refresh: "更新",
        skills_cli_output: "CLI出力",
        skills_install: "インストール",
        skills_install_placeholder: "GitHub ユーザー/リポジトリ",
        skills_install_success: "スキルのインストールが完了しました。",
        skills_remove_success: "スキルが削除されました。",
        skills_installing: "インストール中...",
        skills_confirm_title_install: "スキルのインストール",
        skills_confirm_title_remove: "スキルの削除",
        skills_confirm_desc_install: "[{skill}]をインストールしますか？",
        skills_confirm_desc_remove: "[{skill}]を削除しますか？",

        // Settings Page
        settings_title: "環境設定",
        settings_desc: "エンジンおよびAIサービスの構成を管理します。",
        settings_agent_title: "メインエージェント",
        settings_agent_desc: "デフォルトのモデルとワークスペースを設定します。",
        settings_model: "使用モデル",
        settings_workspace: "ワークスペース",
        settings_providers_title: "APIプロバイダー",
        settings_providers_desc: "AIモデルサービスの認証キーを入力してください。",
        settings_api_key: "APIキー",
        settings_api_base: "APIベース（オプション）",
        settings_security_title: "セキュリティと認証",
        settings_security_desc: "管理者権限を管理します。",
        settings_admin_account: "管理者アカウント",
        settings_change_password: "パスワード変更",
        settings_reset: "設定のリセット",
        settings_save_btn: "設定を保存",
        settings_save_confirm_title: "設定の保存",
        settings_save_confirm_desc: "変更を保存しますか？",
        settings_reset_confirm_title: "設定のリセット",
        settings_reset_confirm_desc: "すべての設定を初期化しますか？",
        settings_save_success: "設定が保存されました。",
        settings_reset_not_impl: "リセット機能はまだ実装されていません。",

        settings_channels_title: "メッセンジャー連携",
        settings_channels_desc: "Discord、Telegramなどの外部連携を構成します。",
        settings_channel_enabled: "有効にする",
        settings_channel_token: "ボットトークン",
        settings_channel_allow_from: "許可するユーザー/ID (カンマ区切り)",

        // Common
        loading: "読み込み中...",
        save: "保存",
        cancel: "キャンセル",
        confirm: "確認",
        delete: "削除",
        status_ok: "システム準備完了",

        // Login
        login_title: "ダッシュボードにアクセスするには管理者パスワードを入力してください",
        admin_password: "管理者パスワード",
        login_btn: "ログイン",
        logging_in: "ログイン中...",
        login_success: "ログインに成功しました",
        login_failed: "パスワードが正しくありません",
        conn_error: "接続エラー",

        // Upgrade
        upgrade_available: "新しいバージョンが利用可能です",
        upgrade_button: "今すぐアップグレード",
        upgrading: "アップグレード中... システムが再起動します。",
        upgrade_confirm: "最新バージョンにアップグレードしますか？",

        // Setup Notice
        setup_notice_title: "初期設定が必要です",
        setup_notice_desc: "MaruBotを正常に動作させるには、少なくとも1つのAIモデルと1つのソーシャルチャネルを設定する必要があります。",
        setup_notice_ai: "AIモデルの設定: {status}",
        setup_notice_channel: "ソーシャルチャネルの設定: {status}",
        setup_notice_configured: "設定済み",
        setup_notice_not_configured: "未設定",
        setup_notice_go_settings: "設定に移動",
    }
};

export const useTranslation = () => {
    const { language } = useLanguageStore();
    return translations[language];
};
