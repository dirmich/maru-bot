# MaruBot 0.9.18

## 주요 변경

- 여러 장비의 marubot에 독립 작업을 병렬로 위임하고 의존 작업은 순서대로 실행합니다.
- Admin에서 전체 장비 상태, 작업·승인·결과, 사용자별 기억과 Wiki, 활동 기록을 관리합니다.
- Google 사용자별 저장소 격리, HTTPS 장비 인증, 로컬 도구 정책과 개별 승인, 중복 실행 방지·취소·복구를 제공합니다.
- 단기·장기 기억과 사용자 소유 장비 사이의 공유 기억을 관리하고, 검토한 기억을 Wiki로 승격합니다.

## 실행

Windows 64비트는 `marubot-windows-amd64.exe`, 32비트는 `marubot-windows-386.exe`를 사용하세요.

```powershell
.\marubot-windows-amd64.exe admin-server
.\marubot-windows-amd64.exe mesh-worker
```

Admin은 기본 `http://localhost:8000`에서 실행됩니다. Google 로그인, 장비별 토큰·TLS·Tailscale을 먼저 설정해야 합니다. 설정 방법은 [환경 설정](https://github.com/dirmich/maru-bot/blob/0.9.18/docs/ENVIRONMENT.md)과 [배포 문서](https://github.com/dirmich/maru-bot/blob/0.9.18/docs/DISTRIBUTION.md)를 참고하세요.

## 검증과 제한

- 전체 Go 테스트·build·vet, Windows 64·32비트 빌드와 버전 확인, 두 아키텍처 실행 파일의 Admin Chrome UI smoke를 통과했습니다. TLS 테스트 worker의 병렬 실행, 사용자 격리, 개별 승인·취소·SSE 재연결도 자동 테스트에 포함됩니다.
- 실제 Google 로그인과 실제 Windows/Linux/RPi 장비 배포는 운영 환경에서 별도 확인해야 합니다.
- macOS DMG는 현재 Windows 빌드 환경에서 생성하지 못해 이번 릴리스에 포함하지 않습니다. 과거 버전의 macOS 파일을 새 버전으로 재사용하지 않습니다.
- 기존 배포 규칙에 따라 Linux/RPi 바이너리는 릴리스 자산에 포함하지 않습니다. 소스 빌드는 지원합니다.
