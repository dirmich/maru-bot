# 구현 체크리스트

- [x] 1. 기획·원문 이관·버전 관리 checkpoint
- [x] 2. 사용자별 저장소·DAG·승인·기억·복구 및 테스트
- [x] 3. 인증·API·SSE·worker·자연어 위임 및 통합 테스트
- [x] 4. Admin 화면·전체 검증·운영 문서
- [ ] [운영 환경 필요] 실제 Google OAuth client와 HTTPS callback 로그인 검증
- [ ] [장비 필요] 실제 Tailscale Windows/Linux/RPi 장비 등록·병렬 작업 검증

## 구현 완료 상세 항목

- [x] Google 사용자별 저장소·세션·장비 자격 자료·CSRF·JWT 서명 검증
- [x] 장비 등록·수정·새 작업 일시중지/재개·토큰 갱신/폐기·heartbeat
- [x] 능력·부하 기반 병렬 분산, 사용자 지명, DAG 순서, 개별 승인·기한
- [x] 원격 도구 정책·파일 경로 격리·소유자 검증·결과 마스킹
- [x] 멱등 실행 journal, 취소, 재시작 unknown 복구·장비 기록 재조회
- [x] backend 사용자별 단기/장기/Wiki, private/shared, 검토 승격·검색·revision·삭제
- [x] 작업 결과의 단기 기억 저장, 대화용 기억 검색·장기 초안 제안
- [x] 현황·장비·작업·승인·기억·활동 기록 Admin frontend/backend
- [x] Go 전체 테스트·vet, Windows 빌드, Linux amd64/arm64 교차 빌드
- [x] Chrome desktop/mobile UI smoke 및 테스트 화면 보관
- [x] 생성 문서를 docs에 보관, 단계별 vup·commit·push
- [ ] [도구 검토 제한] man.sh 복합 시작/중지 실행 검증 (문법 검사 통과, 자동 승인 검토에서 차단)
