# MaruBot - Ultra-lightweight AI Assistant (v0.4.43) 🦞

**MaruBot**은 극강의 효율성을 추구하는 [PicoClaw](https://github.com/sipeed/picoclaw)의 철학을 계승하여, 단 **10MB의 메모리(RAM)** 환경에서도 구동 가능한 초경량 Physical AI 에이전트입니다. Raspberry Pi부터 일반 Linux 서버, Windows PC까지 지원하며, 스스로 능력을 확장하는 '자기 진화' 능력을 갖춘 가장 똑똑한 개인용 비서입니다.

---

## ✨ 핵심 기능 (Key Features)

### 1. 🚀 초경량·고성능 (Ultra-lightweight)
- **10MB RAM:** 최적화된 Go 바이너리를 사용하여 임베디드 장치에서도 부담 없이 동작합니다.
- **Single Binary:** 의존성 없는 단일 바이너리 배포로 설치와 관리가 매우 간편합니다.

### 2. 🌍 멀티 플랫폼 지원 (Multi-Platform)
- **Raspberry Pi:** GPIO, 카메라, 각종 센서 제어 완벽 지원 (ARM32/64).
- **Linux:** Ubuntu, Debian, AWS EC2 등 모든 일반 Linux 환경 지원.
- **Windows:** 64bit 및 32bit 아키텍처 공식 지원 (바이너리 제공).

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

---

## 📂 폴더 구조
- `/config`: MaruBot 하드웨어 및 에이전트 전역 설정.
- `/skills`: AI 에이전트가 학습한 전문 지식 및 가이드라인 (`SKILL.md`).
- `/extensions`: `create_tool`로 생성된 동적 스크립트 도구들.
- `/memory`: SQLite 대화 데이터베이스 및 팩트 보관함.

---

## 🚀 빠른 시작 (Quick Start)

### 1. 원클릭 설치 (Linux/WSL/Git Bash)
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. Windows에서 실행
[공개 저장소 Releases](https://github.com/dirmich/maru-bot/tree/main/releases)에서 본인의 아키텍처에 맞는 `exe` 파일을 다운로드하여 실행하세요.

### 3. 필수 설정 (API 키 등록)
```bash
# OpenAI API 키 설정 예시
marubot config set providers.openai.api_key "YOUR_KEY"

# 기본 모델 선택
marubot config set agents.defaults.model "gpt-4o"
```

### 4. 에이전트 대화 시작
```bash
marubot agent
```
*(또는 `marubot start`를 통해 웹 대시보드 http://localhost:8080 이용 가능)*

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

*Developed & Analyzed by Antigravity AI (2026-03-11)*
