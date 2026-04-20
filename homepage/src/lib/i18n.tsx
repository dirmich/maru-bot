import React, { createContext, useContext, useState, useEffect } from 'react';

export type Language = 'en' | 'ko' | 'ja' | 'es';

export const translations = {
  en: {
    hero_title: "Marubot",
    hero_subtitle: "Ultra-lightweight Personal AI Agent",
    hero_desc: "Your powerful, private, and localized AI companion. Simple to install, easy to use, and supports various messaging channels.",
    features_title: "Key Features",
    feature_1_title: "Privacy First",
    feature_1_desc: "Your data stays on your device. No cloud storage, no data harvesting.",
    feature_2_title: "Multi-Channel",
    feature_2_desc: "Chat via Discord, Telegram, Slack, WhatsApp, or Webhook.",
    feature_3_title: "Lightweight",
    feature_3_desc: "Designed to run on everything from Raspberry Pi to high-end PCs.",
    advantages_title: "Why Marubot?",
    advantages_1: "One-click installation",
    advantages_2: "Low resource usage (under 50MB RAM)",
    advantages_3: "Extensible channel support",
    install_title: "Installation Guide",
    channels_title: "Supported Channels",
    tokens_title: "How to Get Tokens",
    download: "Download",
    get_started: "Get Started",
    win_install: "Windows (x64)",
    win_step_1: "1. Download the latest .exe from GitHub Releases.",
    win_step_2: "2. Run the application (allow through SmartScreen if prompted).",
    win_step_3: "3. Access the Web Admin via tray icon menu.",
    mac_install: "macOS (Intel/Silicon)",
    mac_step_1: "1. Download for your architecture (Intel/Apple Silicon).",
    mac_step_2: "2. Drag to Applications or run installer.",
    mac_step_3: "3. Start Marubot from Launchpad.",
    linux_install: "Linux & Raspberry Pi",
    linux_step_1: "Run the following command in your terminal:",
    shell_command: "Shell Command",
    platform_win: "Windows",
    platform_mac: "macOS",
    platform_linux: "Linux/RPi",
    how_to_telegram: "1. Search for @BotFather in Telegram.\n2. Use /newbot command.\n3. Copy the token and paste it in Marubot settings.",
    how_to_slack: "1. Create an app at api.slack.com.\n2. Enable Socket Mode and Event Subscriptions.\n3. Add necessary scopes and install to your workspace.",
    how_to_discord: "1. Visit Discord Developer Portal.\n2. Create Application -> Bot.\n3. Enable 'Message Content Intent' and copy the token.",
    how_to_whatsapp: "1. Requires a WhatsApp Bridge URL and API Key.\n2. Enter the details in Marubot settings.",
    how_to_webhook: "1. Marubot provides a local webhook endpoint.\n2. Configure your service to send POST requests to this URL.",
    footer_text: "Marubot - Empowering your digital life with AI.",
    close: "Close"
  },
  ko: {
    hero_title: "Marubot (마루봇)",
    hero_subtitle: "초경량 개인용 AI 에이전트",
    hero_desc: "강력하고 안전하며 지역화된 나만의 AI 동반자. 쉬운 설치와 사용법, 다양한 메신저 채널을 지원합니다.",
    features_title: "주요 기능",
    feature_1_title: "프라이버시 최우선",
    feature_1_desc: "데이터는 오직 기기에 저장됩니다. 클라우드 수집이나 서버 저장이 없습니다.",
    feature_2_title: "멀티 채널 연동",
    feature_2_desc: "디스코드, 텔레그램, 슬랙, 왓츠앱, 웹훅 등 다양한 채널로 대화하세요.",
    feature_3_title: "가벼운 성능",
    feature_3_desc: "라즈베리 파이부터 고성능 PC까지 어디서나 부드럽게 작동합니다.",
    advantages_title: "왜 Marubot인가요?",
    advantages_1: "원클릭 설치 및 실행",
    advantages_2: "매우 낮은 리소스 사용량 (50MB 이하)",
    advantages_3: "다양한 메신저 채널 확장성",
    install_title: "설치 가이드",
    channels_title: "지원 채널",
    tokens_title: "토큰 발급 방법",
    download: "다운로드",
    get_started: "시작하기",
    win_install: "Windows 설치 (x64)",
    win_step_1: "1. GitHub Release에서 최신 .exe 파일을 다운로드하세요.",
    win_step_2: "2. 프로그램을 실행하세요. (SmartScreen 경고 시 '실행' 선택)",
    win_step_3: "3. 트레이 아이콘 메뉴에서 웹 관리 도구에 접속하세요.",
    mac_install: "macOS 설치 (Intel/Silicon)",
    mac_step_1: "1. 아키텍처(Intel/Apple Silicon)에 맞는 최신 버전을 받으세요.",
    mac_step_2: "2. 응응 프로그램 폴더로 드래그하거나 설치 프로그램을 실행하세요.",
    mac_step_3: "3. Launchpad에서 Marubot을 실행하세요.",
    linux_install: "Linux 및 Raspberry Pi 설치",
    linux_step_1: "터미널에서 아래 명령어를 실행하세요:",
    shell_command: "쉘 명령어",
    platform_win: "Windows",
    platform_mac: "macOS",
    platform_linux: "Linux/RPi",
    how_to_telegram: "1. 텔레그램에서 @BotFather를 검색하세요.\n2. /newbot 명령어로 봇을 만듭니다.\n3. 발급된 API Token을 마루봇 설정에 입력하세요.",
    how_to_slack: "1. api.slack.com에서 앱을 생성하세요.\n2. Socket Mode 및 Event Subscriptions를 활성화하세요.\n3. 권한(Scope) 추가 후 워크스페이스에 설치하세요.",
    how_to_discord: "1. Discord Developer Portal에 접속하세요.\n2. Application 생성 후 Bot 메뉴로 이동하세요.\n3. 'Message Content Intent' 권한을 활성화하고 토큰을 복사하세요.",
    how_to_whatsapp: "1. WhatsApp 브리지 URL과 API Key가 필요합니다.\n2. 마루봇 설정 메뉴에 해당 정보를 입력하세요.",
    how_to_webhook: "1. 마루봇은 로컬 웹훅 엔드포인트를 제공합니다.\n2. 외부 서비스에서 이 URL로 POST 요청을 보내도록 설정하세요.",
    footer_text: "Marubot - AI와 함께하는 스마트한 디지털 비서.",
    close: "닫기"
  },
  ja: {
    hero_title: "Marubot",
    hero_subtitle: "超軽量パーソナルAIエージェント",
    hero_desc: "強力でプライベート、ローカライズされたAIコンパニオン。インストールが簡単で使いやすく、様々なメッセージングチャネルをサポートします。",
    features_title: "主な機能",
    feature_1_title: "プライバシー重視",
    feature_1_desc: "データはデバイスにのみ保存されます。クラウドへの収集はありません。",
    feature_2_title: "マルチチャネル",
    feature_2_desc: "Discord、Telegram、Slack、WhatsAppなどでチャット可能です。",
    feature_3_title: "軽量設計",
    feature_3_desc: "Raspberry PiからハイエンドPCまで、あらゆる環境で動作します。",
    advantages_title: "なぜMarubotなのか？",
    advantages_1: "ワンクリックでインストール",
    advantages_2: "低リソース消費 (50MB RAM以下)",
    advantages_3: "拡張可能なチャネルサポート",
    install_title: "インストールガイド",
    channels_title: "サポートされているチャネル",
    tokens_title: "トークンの取得方法",
    download: "ダウンロード",
    get_started: "始める",
    win_install: "Windows (x64)",
    win_step_1: "1. GitHubから最新の.exeをダウンロードします。",
    win_step_2: "2. アプリを実行します（SmartScreenが出た場合は許可してください）。",
    win_step_3: "3. トレイアイコンから管理画面에アクセスします。",
    mac_install: "macOS (Intel/Silicon)",
    mac_step_1: "1. CPUに合わせて最新バージョンをダウンロードします。",
    mac_step_2: "2. アプリケーションにドラッグするか、インストーラーを実行します。",
    mac_step_3: "3. Launchpadから起動します。",
    linux_install: "Linux & Raspberry Pi",
    linux_step_1: "ターミナルで以下のコマンドを実行してください：",
    shell_command: "シェルコマンド",
    platform_win: "Windows",
    platform_mac: "macOS",
    platform_linux: "Linux/RPi",
    how_to_telegram: "1. Telegramで@BotFatherを検索します。\n2. /newbotコマンドを使用します。\n3. トークンをコピーして設定に入力します。",
    how_to_slack: "1. api.slack.comでアプリを作成します。\n2. Socket Modeを有効にします。\n3. スコープを追加してインストールします。",
    how_to_discord: "1. Discord Developer Portalにアクセスします。\n2. Botセクションでトークンをコピーします。\n3. Message Content Intentを有効にします。",
    how_to_whatsapp: "1. WhatsAppブリッジURLとAPIキーを入力します。",
    how_to_webhook: "1. ローカルWebhookエンドポイントを設定します。",
    footer_text: "Marubot - AIでデジタルライフをより豊かに。",
    close: "閉じる"
  },
  es: {
    hero_title: "Marubot",
    hero_subtitle: "Agente de IA Personal Ultraligero",
    hero_desc: "Tu compañero de IA potente, privado y localizado. Fácil de instalar, simple de usar y compatible con varios canales de mensajería.",
    features_title: "Características Principales",
    feature_1_title: "Privacidad Primero",
    feature_1_desc: "Tus datos permanecen en tu dispositivo. Sin recolección en la nube.",
    feature_2_title: "Multicanal",
    feature_2_desc: "Chatea por Discord, Telegram, Slack, WhatsApp o Webhook.",
    feature_3_title: "Ligero",
    feature_3_desc: "Diseñado para funcionar desde Raspberry Pi hasta PCs de gama alta.",
    advantages_title: "¿Por qué Marubot?",
    advantages_1: "Instalación en un clic",
    advantages_2: "Bajo consumo de recursos (menos de 50MB RAM)",
    advantages_3: "Soporte de canales extensible",
    install_title: "Guía de Instalación",
    channels_title: "Canales Soportados",
    tokens_title: "Cómo Obtener Tokens",
    download: "Descargar",
    get_started: "Empezar",
    win_install: "Windows (x64)",
    win_step_1: "1. Descarga el .exe más reciente de GitHub Releases.",
    win_step_2: "2. Ejecuta la aplicación (permite en SmartScreen si es necesario).",
    win_step_3: "3. Accede al Web Admin desde el icono de la bandeja.",
    mac_install: "macOS (Intel/Silicon)",
    mac_step_1: "1. Descarga la versión para tu arquitectura (Intel/Apple Silicon).",
    mac_step_2: "2. Arrastra a Aplicaciones o ejecuta el instalador.",
    mac_step_3: "3. Inicia Marubot desde el Launchpad.",
    linux_install: "Linux y Raspberry Pi",
    linux_step_1: "Ejecuta el siguiente comando en tu terminal:",
    shell_command: "Comando de Shell",
    platform_win: "Windows",
    platform_mac: "macOS",
    platform_linux: "Linux/RPi",
    how_to_telegram: "1. Busca @BotFather en Telegram.\n2. Usa el comando /newbot.\n3. Copia el token y pégalo en Marubot.",
    how_to_slack: "1. Crea una app en api.slack.com.\n2. Activa Socket Mode.\n3. Añade permisos e instala.",
    how_to_discord: "1. Visita Discord Developer Portal.\n2. Crea Aplicación -> Bot.\n3. Activa 'Message Content Intent'.",
    how_to_whatsapp: "1. Ingresa la URL del puente y la clave API.",
    how_to_webhook: "1. Configura el endpoint de webhook local.",
    footer_text: "Marubot - Potenciando tu vida digital con IA.",
    close: "Cerrar"
  }
};

interface LanguageContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  t: typeof translations.en;
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

export const LanguageProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [language, setLanguage] = useState<Language>(() => {
    const saved = localStorage.getItem('language');
    return (saved as Language) || 'en';
  });

  useEffect(() => {
    localStorage.setItem('language', language);
  }, [language]);

  const t = translations[language];

  return (
    <LanguageContext.Provider value={{ language, setLanguage, t }}>
      {children}
    </LanguageContext.Provider>
  );
};

export const useTranslation = () => {
  const context = useContext(LanguageContext);
  if (!context) throw new Error('useTranslation must be used within LanguageProvider');
  return context;
};
