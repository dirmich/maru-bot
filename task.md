# 📋 MaruBot 구현 작업 리스트 (Tasks)

본 문서는 MaruBot 프로젝트의 단계별 구현 상태를 추적합니다. 작업 완료 시 상태를 업데이트하고 관련 내용을 커밋합니다.

---

## 🟢 Phase 1: 기반 구축 및 기본 인터랙션 (완료 및 검증)
- [x] PicoClaw 소스 코드 독립화 및 모듈 리팩토링 (`marubot` 모듈)
- [x] Raspberry Pi 하드웨어 설정 자동화 스크립트 작성 (`setup-rpi.sh`)
- [x] CSI/USB 카메라 통합 지원 도구 구현 (`camera_capture`)
- [x] 기본 설정 템플릿 및 프로젝트 구조 수립
- [x] 원클릭 설치 스크립트 (`install.sh`) 구현

---

##  Phase 2: 모빌리티 및 공간 인지 (완료)
- [x] **Task 2.1: PWM 기반 모터 제어 도구 구현**
    - [x] `periph.io` 기반 PWM 라이브러리 연동
    - [x] `move_forward`, `stop`, `set_speed` 등 기본 이동 API 구현
    - [x] 하드웨어 연결 가이드 작성 (README 및 코멘트 반영)
- [x] **Task 2.2: 초음파 센서(HC-SR04) 장애물 감지 도구 구현**
    - [x] 에코/트리거 핀 제어 로직 구현
    - [x] `get_distance` 도구 추가
- [x] **Task 2.3: IMU(MPU6050) 가속도/자이로 센서 연동**
    - [x] I2C 통신 기반 센서 데이터 읽기
    - [x] 로봇의 기울기 및 방향(Bearing) 계산 도구 구현
- [x] **Task 2.4: 시각 기반 위치 추정 및 추적(Vision Tracking)**
    - [x] 표준 라이브러리 기반 경량 이미지 프로세싱 구현
    - [x] 특정 색상 추적 알고리즘 구현 (`track_color` 도구)

---

## 🟢 Phase 3: 공중 기동 및 자율 주행 확장 (완료)
- [x] **Task 3.1: MAVLink 브리지 구현**
    - [x] FC(Flight Controller)와의 시리얼 통신 연동 (Gomavlib v3 기반)
    - [x] 기본 비행 명령(이륙, 착륙, 정지비행) 도구 구현 (`drone_control`)
- [x] **Task 3.2: GPS 및 절대 좌표 기반 이동 구현**
    - [x] NMEA 데이터 파싱 및 현재 위치 도구 구현 (`get_location`)
    - [x] Waypoint 기반 경로 비행 API 추가 (`drone_control` 내 `goto` 명령)
- [x] **Task 3.3: 자율 주행 및 긴급 복구 로직**
    - [x] 자동 복귀(RTL) 및 긴급 강제 착륙(Emergency Stop) 도구 구현
    - [x] 비상 시 하드웨어 제어 인터페이스 구축

---

## 🟢 Phase 4: 자기 진화 및 동적 확장 (완료)
- [x] **Task 4.1: 실시간 도구 생성 시스템 구축**
    - [x] AI가 스스로 스크립트를 작성하고 도구로 등록하는 `create_custom_tool` 도구 구현
    - [x] `/extensions` 디렉토리를 통한 동적 도구 로드 메커니즘 구축
    - [x] 무중단 기능 확장 및 런타임 도구 업데이트 검증

---

## 🟢 Phase 5: 맞춤형 설정 및 사용자 편의성 (완료)
- [x] **Task 5.1: 설정 파일 분리 및 관리 도구 구현**
    - [x] 사용자 설정을 기록하는 `usersetting.json` 및 오버라이드 시스템 구축
    - [x] 에이전트 내에서 설정을 조회하고 수정하는 `config` 도구 추가
    - [x] 명령줄 인터페이스(CLI)용 `marubot config` 명령어 구현
- [x] **Task 5.2: 공개 레포지토리 배포 자동화 구축**
    - [x] 비공개 환경에서 공개용 파일만 선별하여 동기화하는 `make public` 타겟 구현
    - [x] 공개 배포 시 모든 명칭 자동 치환 (`marubot` -> `marubot`) 엔진 구축
    - [x] MIT 라이선스 파일 추가 및 공개 배포 가이드 수립

---

## ✅ 작업 완료 로그 (Log)
- **2026-02-10**: 프로젝트 독립화, 리팩토링 및 카메라 통합 도구 구현 완료. (Step 1~4 완료)
- **2026-02-11**: GitHub Gist 기반 설치 스크립트 적용 및 설치 가이드 개선.
- **2026-02-11**: 사용자 설정 분리(`usersetting.json`) 및 `config` 관리 도구 구현 완료. (Phase 5)
- **2026-02-11**: 공개 배포 자동화(`make public`) 및 전면 명칭 치환 프로세스 완성.
- **2026-02-13**: 배포 스크립트(`publish.sh`)의 `.git` 폴더 보존 로직 강화 (삭제 방지).
