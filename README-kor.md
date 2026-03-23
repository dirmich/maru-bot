<<<<<<< HEAD
# MaruBot - Ultra-lightweight AI Assistant (0.4.86) 🦞
=======
# MaruBot - Ultra-lightweight AI Assistant (0.4.87) 🦞
>>>>>>> 42feaf4 (슬랙 연동 로깅 개선 및 0.4.87 릴리즈)

**MaruBot**은 극강의 효율성을 추구하며, 단 **10MB의 메모리(RAM)** 환경에서도 구동 가능한 초경량 Physical AI 에이전트입니다. Raspberry Pi부터 일반 Linux 서버, Windows PC까지 지원하며, 스스로 능력을 확장하는 '자기 진화' 능력을 갖춘 가장 똑똑한 개인용 비서입니다.

---

## ✨ 핵심 기능 (Key Features)

### 1. 🚀 초경량·고성능 (Ultra-lightweight)
- **10MB RAM:** 최적화된 Go 바이너리를 사용하여 임베디드 장치에서도 부담 없이 동작합니다.
- **Single Binary:** 의존성 없는 단일 바이너리 배포로 설치와 관리가 매우 간편합니다.

### 2. 🌍 멀티 플랫폼 지원 (Multi-Platform)
- **Raspberry Pi:** GPIO, 카메라, 각종 센서 제어 완벽 지원 (ARM32/64).
- **Linux:** Ubuntu, Debian, AWS EC2 등 모든 일반 Linux 환경 지원.
- **Windows:** 64bit 및 32bit 아키텍처 공식 지원 (바이너리 제공).
- **macOS:** Intel 및 Apple Silicon(M1/M2/M3) 공식 지원 (DMG 제공).

### 3. 🧬 자기 진화 엔진 (Auto-Evolution)
- **`create_tool`**: 새로운 원자적 기능을 담당하는 Bash/Python 스크립트를 스스로 코딩하고 즉시 도구로 등록합니다.
- **`create_skill`**: 복잡한 워크플로우나 지침을 담은 상위 수준의 '스킬'을 폴더 기반으로 자동 생성합니다.

### 4. 🧠 스마트 기억 시스템 (RAG Memory)
- **SQLite FTS5:** SQLite와 Full-Text Search를 활용하여 수만 개의 대화 내역에서도 관련 정보를 1초 내에 검색하여 대화에 반영합니다.
- **Facts & Preferences:** 사용자의 성향과 중요한 사실들을 영구적으로 기억합니다.

### 5. 🛠️ 강력한 자동화 및 제어 도구
- **Cron 예약 작업:** "내일 오전 9시에 날씨 알려줘"와 같은 예약 명령을 수행합니다.
- **SSH 매니저:** 원격 서버 접속을 위한 공개키 생성 및 배포를 지능적으로 관리합니다.
- **MAVLink Drone:** 드론 컨트롤러와 연동하여 물리적 비행 제어가 가능합니다.
31: 
### 6. 🌐 중앙 관리 및 홈페이지 (Admin & Landing)
- **공식 홈페이지:** 멀티링구얼(KO, EN, JA, ES) 지원 및 원클릭 설치 가이드 제공.
- **중앙 관리(Admin):** Google SSO 연동으로 여러 Marubot 인스턴스의 상태(OS, 메모리 등)를 한눈에 파악.

---

### 🤖 AI Models & Fallback
MaruBot supports multiple AI providers. You can configure a primary model and a list of fallback models in your `config.json`:

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
If the primary provider fails, MaruBot will automatically try the fallback models in order.

### 🪟 Windows Deployment
For Windows users, we provide two types of distributions in the [Releases](https://github.com/dirmich/maru-bot/releases) section:
1. **Single Binary (`marubot.exe`)**: A standalone executable for quick use.
2. **Installable Package (`marubot-windows-x64.zip`)**: Includes the executable, default configuration, and a quick-start guide.

> [!TIP]
> **Windows 사용자 참고**: 상업용 인증서로 서명되지 않은 바이너리이므로, 실행 시 SmartScreen의 'Windows의 PC 보호' 경고가 나타날 수 있습니다. **'추가 정보'**를 클릭한 후 **'실행'** 버튼을 선택하여 진행해 주세요.

> [!TIP]
> **macOS 사용자 참고**: 애플의 공증(Notarization) 절차를 거치지 않은 버전의 경우 처음 실행 시 '악성 코드가 없음을 확인할 수 없습니다'라는 경고가 뜰 수 있습니다. 공식적으로 서명된 버전은 바로 실행 가능하지만, 그렇지 않은 경우 **'시스템 설정 > 개인정보 보호 및 보안'**에서 **'확인 없이 열기'**를 클릭하거나 앱 아이콘에서 **'우측 클릭 > 열기'**를 선택해 주세요.

## 📂 폴더 구조
- `/config`: MaruBot 하드웨어 및 에이전트 전역 설정.
- `/skills`: AI 에이전트가 학습한 전문 지식 및 가이드라인 (`SKILL.md`).
- `/extensions`: `create_tool`로 생성된 동적 스크립트 도구들.
- `/memory`: SQLite 대화 데이터베이스 및 팩트 보관함.

---

## 🚀 빠른 시작 (Quick Start)

### 1. 🐧 Linux / 🍎 macOS (Terminal)
터미널에서 아래 명령어를 실행하여 즉시 설치할 수 있습니다 (curl 필요):
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. 🪟 Windows (GUI/Manual)
Windows 사용자는 터미널 명령어보다는 **공식 릴리스 페이지**에서 파일을 다운로드하여 실행하는 것을 권장합니다:
1. [공식 릴리스 페이지(Releases)](https://github.com/dirmich/maru-bot/releases)에 접속합니다.
2. 본인의 OS(64bit 또는 32bit)에 맞는 `marubot-windows-xxx.zip` 또는 `exe` 파일을 다운로드합니다.
3. 다운로드한 파일을 실행하면 자동으로 `~/.marubot/bin` 폴더에 설치되고 트레이 아이콘이 활성화됩니다.

### 3. 🍎 macOS (GUI)
1. [공식 릴리스 페이지](https://github.com/dirmich/maru-bot/releases)에서 본인의 CPU(Intel 또는 Apple Silicon)에 맞는 `.dmg` 파일을 다운로드합니다.
2. DMG 파일을 열고 `MaruBot.app`을 실행하면 메뉴 막대(트레이)에 아이콘이 나타납니다.

### 3. 필수 설정 (API 키 및 기본 모델 등록)
```bash
# OpenAI API 키 설정 예시
marubot config set providers.openai.api_key "YOUR_KEY"

# 기본 모델 선택 (예: gpt-4o, gemini-2.5-flash 등)
marubot config set agents.defaults.model "gpt-4o"
```
*💡 기본 모델 연결이 실패하더라도 다른 프로바이더의 API 키가 존재하면 자동으로 Fallback 됩니다!*

### 4. 에이전트 대화 시작
```bash
marubot agent
```
*(또는 `marubot start`를 통해 웹 대시보드 http://localhost:8080 이용 가능)*

---

## 🧩 확장 가이드 (Skills & Tools)
MaruBot은 사용자가 쉽게 기능을 확장할 수 있습니다.

### 도구(Tool) 추가하기
단일 스크립트 실행 등 단순한 동작을 등록하려면 AI에게 직접 요청하세요.
```text
"시스템 정보를 보여주는 파이썬 스크립트를 작성해서 새로운 도구로 등록해줘"
```
AI가 스스로 **`extensions` 폴더**에 스크립트와 메타데이터를 저장하고 런타임에 즉시 활용합니다.

### 스킬(Skill) 추가하기
복합적인 워크플로우나 프롬프트 템플릿 제어가 필요하다면 스킬을 제작합니다.
```text
"새 코딩 템플릿 스킬을 만들어줘"
marubot skills show <skill-name>
```
물리적으로 `~/.marubot/workspace/skills/<스킬명>` 디렉토리를 만들어 그 안에 `SKILL.md`를 작성하면 자동 연동됩니다.

---

## 🛠️ 하드웨어 연동 가이드
MaruBot은 구동 플랫폼을 감지하여 사용 가능한 도구를 자동 선별합니다.
- **GPIO**: LED, 버튼, 릴레이 제어.
- **Camera**: Libcamera 또는 USB 카메라를 통한 시각 인식.
- **Sensors**: MPU6050(IMU), 초음파 센서, GPS 등 연동.
- **Motor**: 서보 및 DC 모터 직접 제어.

---

## 📝 라이선스
MaruBot은 MIT 라이선스를 따릅니다. 사용자는 자유롭게 수정하고 배포할 수 있습니다.

*Developed & Analyzed by Antigravity AI (2026-03-20)*
