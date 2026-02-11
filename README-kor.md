# 🤖 MaruBot (마루 미니봇)

**MaruBot**은 MaruBot의 초경량 엔진을 기반으로, Raspberry Pi와 같은 SBC(Single Board Computer)에서 하드웨어를 직접 제어하고 소통하기 위해 최적화된 **"Physical AI Assistant"**입니다.

---

## ✨ 핵심 컨셉
1. **MaruBot 엔진 재사용**: MaruBot의 고효율 Go 바이너리를 그대로 사용하여 10MB 이하의 RAM 점유율을 유지합니다.
2. **Raspberry Pi 최적화**: GPIO, 카메라, 마이크, 스피커 권한 설정을 자동화합니다.
3. **하이퍼-로컬 설정**: 복잡한 JSON 편집 대신 전용 스크립트(`maru-setup.sh`)를 통해 대화형으로 설정을 완료합니다.
4. **물리적 상호작용**: AI 에이전트가 서보 모터, LED, 각종 센서(DHT, PIR 등)를 제어할 수 있는 도구가 사전 포함되어 있습니다.

---

## 📂 폴더 구조
- `/config`: 마루 미니봇 전용 하드웨어 및 에이전트 설정 파일
- `maru-setup.sh`: 라즈베리 파이 초기화 및 하드웨어 연동 자동화 스크립트
- `/tools`: AI 에이전트가 사용할 GPIO/I2C/SPI 제어 유틸리티 (구현 예정)
- `/bin`: MaruBot 바이너리 링크 또는 실행 파일 보관

---

## 🚀 빠른 시작 (Quick Start)

### 1. 원클릭 설치 (GitHub Gist 권장)
가장 빠르고 간편한 설치 방법입니다. 본인의 Gist에서 **Raw** 버튼을 눌러 얻은 URL을 사용하세요:

```bash
# Official MaruBot One-Line Installer
curl -fsSL https://gist.githubusercontent.com/dirmich/367961d107d6e0f35f1c3156dc55f7d5/raw/install.sh | bash
```

#### 💡 나만의 설치 Gist 만드는 법:
1. [gist.github.com](https://gist.github.com/) 접속 (Secret/Public 상관 없음)
2. 파일명을 `install.sh`로 입력하고 본 프로젝트의 `install.sh` 내용 붙여넣기
3. 생성 후 페이지 우측 상단의 **Raw** 버튼 클릭
4. 이동된 페이지의 주소(URL)를 복사하여 `curl -fsSL <복사한_URL> | bash` 명령어로 사용

### 2. 수동 설치 및 하드웨어 준비
만약 위 명령어가 작동하지 않거나 수동 설치를 원할 경우:
1. Go 1.24+ 및 필수 도구 설치 (`sudo apt install -y git make golang libcamera-apps`)
2. 리포지토리 클론: `git clone https://github.com/dirmich/maru-bot.git`
3. 설정 스크립트 실행: `cd marubot && bash maru-setup.sh`
이 스크립트는 다음을 수행합니다:
- Raspberry Pi GPIO 라이브러리(`/dev/gpiomem`) 권한 확인
- 카메라 및 오디오 인터페이스 활성화 여부 점유
- 전용 대화형 설정 위저드 실행

### 3. 에이전트 실행
```bash
./maru-run.sh
```

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

*개발 및 분석: Antigravity AI (2026-02-10)*
