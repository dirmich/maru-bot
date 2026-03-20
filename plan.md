# 🤖 MaruBot (마루 미니봇) 프로젝트 상세 기획서

## 1. 프로젝트 개요 (Project Overview)
**MaruBot**은 초경량 AI 엔진 기술을 독자적으로 구축하여, Raspberry Pi와 같은 SBC(Single Board Computer) 환경에서 물리적 세계와 소통하도록 설계된 **"전천후 피지컬 AI 어시스턴트"**입니다. 단순한 텍스트 기반 비서를 넘어, 하드웨어 센서와 액추에이터를 직접 제어하여 현실 세계의 문제를 해결하는 것을 목표로 합니다.

---

## 2. 핵심 목표 (Vision & Goals)
1. **극강의 리소스 효율성**: 10MB 미만의 메모리 점유율을 유지하여 저사양 하드웨어에서도 상시 가동 가능.
2. **피지컬 컴퓨팅의 AI화**: GPIO, I2C, SPI 등 복잡한 하드웨어 제어를 AI 에이전트가 자연어로 수행.
3. **높은 범용성 및 확장성**: 바퀴(이동형), 날개(드론형) 등 모빌리티 확장이 용이한 환경 제공.
4. **설치 및 환경 설정의 압도적 편의성**: 원클릭 스크립트를 통한 복잡한 리눅스/하드웨어 권한 설정 자동화.

---

## 3. 주요 하드웨어 지원 범위 (Hardware Support)
- **메인 컴퓨팅**: Raspberry Pi (모든 모델), RISC-V SBC, ARM64 기반 리눅스 기기.
- **이미지 장치**: Raspberry Pi 전용 CSI 카메라(libcamera), 범용 USB 웹캠(fswebcam/ffmpeg).
- **오디오 기기**: USB 마이크/스피커, I2S DAC/ADC (ALSA 인터페이스).
- **센서/액추에이터**: GPIO 핀, I2C 센서(상태 모니터링), PWM 기반 서보/DC 모터 (모빌리티).

---

## 4. 단계별 개발 로드맵 (Development Roadmap)

### **Phase 1: 기반 구축 및 기본 인터랙션 (현재 단계)**
- [x] 독자적인 독립 엔진(`marubot`) 구축 및 모듈 리팩토링.
- [x] Raspberry Pi 전용 하드웨어 설정 자동화 스크립트(`setup-rpi.sh`) 개발.
- [x] 복합 카메라 지원 툴(`camera_capture`) 통합 (CSI & USB).
- [x] GPIO 제어를 위한 기본 도구 및 설정 템플릿 제공.

### **Phase 2: 모빌리티 및 위치 파악 (Mobility & Positioning)**
- **바퀴 달린 마루**: DC 모터 및 L298N 등 드라이버 제어를 위한 PWM 도구 추가.
- **공간 인지**: 초음파 센서(HC-SR04) 및 IMU(MPU6050) 연동을 통한 장애물 회피 구현.
- **사용자 추적**: 카메라 이미지 분석을 통해 사용자를 인식하고 적정 거리를 유지하며 따라가는 기능.

### **Phase 3: 공중 기동 및 자율 주행 (Aerial & Autonomy)**
- **드론 확장**: MAVLink 프로토콜을 통한 비행 제어 장치(FC) 연동 도구 구현.
- **고급 비행 제어**: GPS 및 기압계를 활용한 자동 고도 유지 및 경로 이동(Waypoint) 명령 지원.
- **에지 AI 고도화**: 인터넷 연결 없이도 로컬에서 기본적인 음성 인식 및 상황 판단 수행.

---

## 5. 기술 아키텍처 (Technical Architecture)

### **A. Software Stack**
- **Language**: Go (정적 바이너리 배포로 의존성 제로화)
- **Engine**: 독립된 `marubot` 패키지 관리
- **Interface**: 
    - 메시징: Telegram, Discord, CLI
    - 하드웨어: `periph.io` (GPIO/I2C), `libcamera` (Camera), `ALSA` (Audio)

### **B. Hardware Interface**
- AI 가 도구(Tool)를 호출하면 쉘 스크립트나 Go 라이브러리를 통해 즉시 물리 장치 제어.
- 샌드박스 환경을 고려한 보안 정책 기반의 하드웨어 접근 제어.

---

## 6. 권장 활용 사례 (Use Cases)
1. **스마트 홈 AI 서비스**: "거실 온도 체크하고 너무 높으면 창문 열어줘" 등의 복합 작업 수행.
2. **교육용 로봇 플랫폼**: AI 알고리즘과 피지컬 컴퓨팅을 동시에 학습할 수 있는 오픈소스 교구.
3. **자율 순찰 및 보안**: 정해진 시간에 집안을 이동하며 카메라로 이상 징후를 감지하고 메신저로 보고.
4. **AI 드론 연구**: 고수준의 AI 로직은 마루 미니봇이 담당하고 비행은 FC가 담당하는 협업 시스템 연구.

---
*기획 일시: 2026-02-10*
*작성자: Antigravity AI*
# 작업 계획 (Implementation Plan)

사용자의 지시에 따라, 항상 수정하기 전에 계획을 세우고 허락을 받은 뒤 작업을 진행합니다.

## 수정 목표
- `marubot dashboard` 명령을 `marubot start`로 변경하고 백그라운드 구동 지원
- `marubot reload` 명령 추가: 실행 중인 설정을 다시 읽어오도록 구성 (데몬 재시작)
- Raspberry Pi (Linux) 환경에서 디바이스 재부팅 시에도 봇이 계속 실행되도록 `systemd` 서비스 등록 및 관리 코드 도입

## 진행 단계

### 1단계: 기능 구현 (진행된 내용 보완 및 점검)
- `cmd/marubot/main.go` 파일 내 `dashboardCmd`를 `startCmd`로 갱신 (완료됨)
- `reloadCmd` 신규 작성 및 설정 리로드 로직(또는 재시작 로직) 구현 (완료됨)
- Linux 환경에서는 `systemd` service 파일(`.config/systemd/user/marubot.service`)을 자동 등록/시작하도록 `installAndRunSystemdService` 함수 구현 (완료됨)
- Windows 환경 등에서는 기존 데몬 모드처럼 프로세스 PID를 종료 후 다시 `start` 명령으로 재실행하도록 처리

### 2단계: 문서 및 스크립트 갱신
- 각종 언어별 `README.md`, `install.sh` 내 `dashboard` 명령어 안내를 `start`로 모두 교체 (부분 완료됨, 최종 점검)
- `task.md` 에 관련된 개발 진행 상황 업데이트

### 3단계: 버전 업 (Version Bump)
- `cmd/marubot/main.go` 내의 `version` 문자열 수정 (예: `0.3.10` -> `0.3.11` 등)
- 필요시 `web-admin/package.json` 버전 맞춤

### 4단계: 소스코드 커밋 및 푸시 (Git)
- `chcp 65001` 실행하여 한글 인코딩 보호
- 수정한 파일들(`main.go`, 각종 `README*`, `install.sh`, `task.md`)을 `git add .` 및 `git commit -F COMMIT_MSG`로 커밋 후 `git push`

### 5단계: 배포 (Publish)
- `scripts/publish.sh` 스크립트 등을 실행하여 수정된 버전을 퍼블리싱 (배포)

> **리뷰 요청**: 이 단계별 진행 과정에 대해 허락해 주시면, 나머지 2~5단계를 순차적으로 진행하겠습니다.
