# 빌드와 장비 연결

## 빌드

Go 1.25 이상에서 저장소 루트에서 실행한다.

```powershell
go test ./...
go build -o build/maru-admin.exe ./admin/backend
go build -o build/marubot.exe ./cmd/marubot
```

frontend는 별도 npm 설치나 빌드 없이 Admin 바이너리에 포함된다. 소스는 `admin/frontend/`, backend 진입점은 `admin/backend/`다. 기존 `cmd/marubot/dashboard/dist` 산출물은 덮어쓰지 않는다.

GitHub Release에서 받은 marubot 실행 파일은 `marubot admin-server` 명령으로 동일한 Admin backend/frontend를 실행한다. 별도 maru-admin 빌드는 선택 사항이다. worker는 같은 실행 파일의 `mesh-worker` 명령을 사용한다.

## 처음 연결하기

1. docs/ENVIRONMENT.md에 따라 Google OAuth·Admin HTTPS·데이터 경로를 설정한다.
2. `build/maru-admin.exe`를 저장소 루트에서 실행한다. `go run ./admin/backend`도 가능하다.
3. 공개 Admin 주소로 접속해 Google로 로그인한다.
4. 장비 탭에서 장비 이름과 worker의 HTTPS 주소를 등록한다. 토큰·장비 ID·소유자 ID를 안전한 환경에 복사한다.
5. 장비에 동일 버전의 marubot을 설치하고 worker 환경변수·TLS·도구 정책을 설정한다.
6. 장비에서 `marubot mesh-worker`를 실행한다. 연결되면 Admin 장비 탭에서 15초 이내 heartbeat와 도구가 표시된다.
7. 서로 다른 두 장비에 `system_status` 작업을 각각 지정하고 승인 수준을 읽기 전용으로 설정한다. 의존 단계는 비워 병렬 실행한다.
8. 결과에서 각 단계의 장비와 출력을 확인한다. shell 또는 쓰기 작업은 개별 승인 후 실행한다.

실제 장비 연결은 해당 장비의 Tailscale, 유효한 인증서, 실행 권한이 필요하다. 운영 서버에 자동 업로드·재시작·릴리스를 수행하는 명령은 이 절차에 포함하지 않는다.

## 데이터와 복구

- `admin-data/auth.sqlite`: 세션, 해시된 토큰 식별자, 암호화된 장비 접속 자료.
- `admin-data/master.key`: AES-GCM 키. 데이터베이스와 함께 백업하되 접근 권한을 제한한다. Windows에서는 해당 서비스 계정만 읽도록 NTFS ACL도 설정한다.
- `admin-data/users/<SHA256(Google sub)>/state.sqlite`: 사용자별 장비, DAG·이벤트·기억·Wiki와 revision.
- worker의 marubot 홈 아래 `worker-state/users/<hash>/state.sqlite`: 장비 실행 journal.
- SQLite WAL 사용 중에는 온라인 backup API 또는 서비스를 안전하게 중지한 상태에서 DB·WAL을 일관되게 백업한다. 파일 하나만 임의 복사하지 않는다.

재시작 시 running/cancelling은 unknown으로 복구한다. Admin 결과에 실행 확인 필요가 표시되면 실제 장비 실행 journal과 부작용을 확인한 후 새 멱등 키의 작업을 제출한다. 같은 작업을 다른 장비에서 자동 재실행하지 않는다. 취소는 이미 수행한 OS 변경을 되돌리지 않는다.

## 운영 확인

Google 실제 로그인, Windows/Linux/RPi 각각의 OS 도구 동작, 실제 Tailscale 장비간 연결은 운영 자격 자료와 장비가 필요하다. 자동 테스트의 TLS worker 검증과 별도로 수행한다. marumesh는 아직 구현하지 않았으며 전송 인터페이스의 후속 구현 대상으로 유지한다.
