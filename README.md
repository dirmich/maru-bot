# 🤖 MaruBot (마루봇)

**MaruBot**은 MaruBot의 초경량 엔진을 기반으로, Raspberry Pi와 같은 SBC(Single Board Computer)에서 하드웨어를 직접 제어하고 소통하기 위해 최적화된 **"Physical AI Assistant"**입니다.

## ✨ 0.4.0 업데이트
- **Webhook 채널 추가**: 외부 서비스에서 HTTP 요청을 통해 마루봇과 실시간 대화가 가능합니다. (동기적 응답 지원)
- **AI 자기 진화 (Self-Evolution)**: AI 에이전트가 스스로 새로운 기술을 설치하거나 도구를 생성한 뒤, 자신을 재시작(`reload`)하여 기능을 확장할 수 있습니다.
- **시스템 제어 도구 (`system_control`)**: AI가 직접 마루봇의 상태를 확인하고, 기술 설치 및 프로세스 재시작을 수행할 수 있는 역량을 갖췄습니다.

## ✨ 0.3.2 업데이트
- **자가 업그레이드(Self Upgrade)**: `marubot upgrade` 명령어로 간편하게 최신 버전으로 업데이트할 수 있습니다.
- **Web Admin 통합**: 더 이상 별도의 Node.js 설치가 필요 없습니다. Web Admin이 Go 바이너리에 내장되어 단일 파일로 실행됩니다.

---

## ✨ 핵심 컨셉
1. **MaruBot 엔진 재사용**: MaruBot의 고효율 Go 바이너리를 그대로 사용하여 10MB 이하의 RAM 점유율을 유지합니다.
2. **Raspberry Pi 최적화**: GPIO, 카메라, 마이크, 스피커 권한 설정을 자동화합니다.
3. **하이퍼-로컬 설정**: 복잡한 JSON 편집 대신 전용 스크립트(`maru-setup.sh`)를 통해 대화형으로 설정을 완료합니다.
4. **물리적 상호작용**: AI 에이전트가 서보 모터, LED, 각종 센서(DHT, PIR 등)를 제어할 수 있는 도구가 사전 포함되어 있습니다.

---

## 📂 폴더 구조
- `/config`: 마루봇 전용 하드웨어 및 에이전트 설정 파일
- `maru-setup.sh`: 라즈베리 파이 초기화 및 하드웨어 연동 자동화 스크립트
- `/tools`: AI 에이전트가 사용할 GPIO/I2C/SPI 제어 유틸리티 (구현 예정)
- `/bin`: MaruBot 바이너리 링크 또는 실행 파일 보관

---

##  사전 준비 (Prerequisites)

시작하기 전에 다음 사항이 준비되었는지 확인하세요:
- **Hardware**: Raspberry Pi (ARM64/32 완벽 지원), 전원 아답터, SD 카드
- **OS**: Raspberry Pi OS (Bullseye 이상 권장)
- **API Key**: OpenAI, Gemini 등 사용할 LLM 서비스의 API 키

---

## 🚀 빠른 시작 (Quick Start)

가장 빠르고 간편하게 마루봇을 시작하는 방법입니다.

### 1. 원클릭 설치
터미널에서 아래 명령어를 실행하여 엔진과 웹 관리자를 한 번에 설치합니다:

```bash
# Official MaruBot One-Line Installer
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. 필수 설정 (API 키 등록)
설치 완료 후, 사용할 AI 모델의 API 키를 등록합니다:

```bash
# OpenAI API 키 설정 예시
marubot config set providers.openai.api_key "YOUR_OPENAI_KEY"

# 기본 모델 선택
marubot config set agents.defaults.model "gpt-4o"
```

### 3. 에이전트 실행
```bash
# 콘솔 대화 모드
marubot agent

# 또는 웹 관리자 대시보드 (http://localhost:8080)
marubot start
```

### 4. 업데이트 (Upgrade)
새로운 기능이 출시되었을 때, 다음 명령어로 간편하게 업데이트하세요:
```bash
marubot upgrade
```
이 명령어는 자동으로 기존 프로세스를 종료하고 최신 버전을 설치한 후 완료 메시지를 표시합니다.


---

## 🛠️ 상세 설치 및 하드웨어 연동 (Detailed Installation)

원클릭 설치가 작동하지 않거나 수동 설정을 원하는 경우:

1.  **필수 도구 설치**: `sudo apt install -y git make golang libcamera-apps`
2.  **리포지토리 클론**: `git clone https://github.com/dirmich/maru-bot.git marubot` (또는 `dirmich/maru-bot`)
3.  **설치 스크립트 실행**: `cd marubot && bash install.sh`
    -   이 스크립트는 Web Admin 빌드, Go 바이너리 임베딩 및 빌드, 리소스 배포를 자동으로 수행합니다.

---

## ⚙️ 설정 (Configuration)

설치가 완료되면 AI 모델을 사용하기 위해 API 키를 설정해야 합니다.

1. **명령줄 도구 사용 (권장)**:
   ```bash
   # OpenAI API 키 설정
   marubot config set providers.openai.api_key "YOUR_KEY"
   
   # 기본 모델 변경
   marubot config set agents.defaults.model "gpt-4o"
   ```

2. **설정 파일 직접 수정**:
   ```bash
   nano ~/.marubot/config.json
   ```
   `providers` 섹션에서 사용할 서비스(openai, gemini 등)의 `api_key` 아래에 본인의 키를 입력합니다.

---

## 🔧 주요 하드웨어 제어 기능
- **GPIO**: LED 제어, 버튼 입력 감지
- **I2C/SPI**: 온도, 습도, 조도 센서 데이터 실시간 읽기
- **Camera**: AI가 직접 현장을 촬영하고 상황 분석 (Libcamera 연동)
- **Audio**: 로컬 마이크를 통한 음성 명령 수신 및 스피커 출력

---

## 📝 라이선스
MaruBot의 철학을 계승하여 MIT License를 따릅니다.

MaruBot은 [picoclaw](https://github.com/sipeed/picoclaw)를 기본으로 Raspberry Pi에 맞게 기능이 추가되었습니다.

*개발 및 분석: Antigravity AI (2026-02-19)*
