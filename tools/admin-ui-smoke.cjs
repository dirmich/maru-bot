// UI는 테스트용 응답으로만 검증한다. 운영 서버에 로그인 우회 기능을 추가하지 않는다.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || "playwright");
const { spawn } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const assert = require("node:assert/strict");

(async () => {
  const root = path.resolve(__dirname, "..");
  const data = fs.mkdtempSync(path.join(os.tmpdir(), "marubot-ui-"));
  const child = spawn(
    process.env.MARUBOT_UI_BINARY ||
      path.join(
        root,
        "build",
        process.platform === "win32" ? "maru-admin.exe" : "maru-admin",
      ),
    process.env.MARUBOT_UI_COMMAND ? [process.env.MARUBOT_UI_COMMAND] : [],
    {
      cwd: root,
      windowsHide: true,
      env: {
        ...process.env,
        MARUBOT_ADMIN_ADDR: "127.0.0.1:18180",
        MARUBOT_ADMIN_PUBLIC_URL: "http://localhost:18180",
        MARUBOT_ADMIN_DATA: data,
        GOOGLE_CLIENT_ID: "",
        GOOGLE_CLIENT_SECRET: "",
        MARUBOT_ADMIN_CERT: "",
        MARUBOT_ADMIN_KEY: "",
        MARUBOT_CONTROL_CA: "",
        MARUBOT_ADMIN_LLM_CONFIG: "",
      },
      stdio: "pipe",
    },
  );
  let browser;
  const errors = [];
  child.stderr.on("data", (b) => errors.push(b.toString()));
  try {
    const base = "http://localhost:18180";
    for (let i = 0; i < 80; i++) {
      try {
        if ((await fetch(base + "/api/auth/status")).ok) break;
      } catch {}
      await new Promise((r) => setTimeout(r, 150));
    }
    browser = await chromium.launch({ channel: "chrome", headless: true });
    const page = await browser.newPage({
      viewport: { width: 1440, height: 1000 },
    });
    const jsErrors = [];
    page.on("pageerror", (e) => jsErrors.push(e.message));
    await page.goto(base);
    await page
      .getByText("Google 로그인 설정이 필요합니다.", { exact: false })
      .waitFor();
    assert.equal((await fetch(base + "/api/overview")).status, 401);
    const now = new Date().toISOString();
    const nodes = [
      {
        id: "windows",
        name: "Windows 사무실 · UI 검증",
        url: "https://100.100.0.1:8443",
        platform: "windows",
        arch: "amd64",
        version: "test",
        protocol: 1,
        tools: ["system_status", "shell"],
        last_seen: now,
        online: true,
        cpu: 18.3,
        memory: 46.7,
      },
      {
        id: "linux",
        name: "Linux 서버 · UI 검증",
        url: "https://100.100.0.2:8443",
        platform: "linux",
        arch: "arm64",
        version: "test",
        protocol: 1,
        tools: ["system_status", "shell"],
        last_seen: now,
        online: true,
        cpu: 7.9,
        memory: 29.4,
      },
    ];
    let memories = [
      {
        id: "memory",
        title: "테스트용 작업 절차",
        content: "UI 검증 데이터입니다. 운영 기억이 아닙니다.",
        kind: "short",
        visibility: "shared",
        revision: 1,
        created_at: now,
        updated_at: now,
        expires_at: new Date(Date.now() + 86400000).toISOString(),
      },
    ];
    let runs = [
      {
        id: "test-run",
        title: "UI 검증 · 두 장비 상태 확인",
        status: "need-approval",
        created_at: now,
        deadline: new Date(Date.now() + 3600000).toISOString(),
        tasks: [
          {
            id: "read",
            instruction: "시스템 현황 조회",
            node_id: "windows",
            tool: "system_status",
            args: {},
            status: "succeeded",
            output: "UI 검증 결과",
          },
          {
            id: "build",
            instruction: "서버 빌드",
            node_id: "linux",
            tool: "shell",
            args: { command: "go build ./..." },
            status: "need-approval",
            approval_level: "local-write",
            depends_on: ["read"],
          },
        ],
      },
    ];
    await page.route("**/api/**", async (route) => {
      const req = route.request();
      const url = new URL(req.url());
      const p = url.pathname;
      let body = {};
      const method = req.method();
      if (p === "/api/me")
        body = {
          owner: "test-owner",
          name: "UI 검증 사용자",
          role: "human",
          csrf: "test-csrf",
        };
      else if (p === "/api/overview")
        body = {
          nodes,
          runs,
          online: nodes.filter((n) => n.online).length,
          running: 0,
          pending: runs
            .flatMap((r) => r.tasks)
            .filter((t) => t.status === "need-approval").length,
          memory_count: memories.length,
          updated_at: now,
        };
      else if (p === "/api/events") {
        if (req.headers().accept?.includes("text/event-stream")) {
          return route.fulfill({
            contentType: "text/event-stream",
            body: "retry: 600000\n\n: fixture\n\n",
          });
        }
        body = [
          { id: 1, type: "ui.test", message: "UI 테스트용 활동", time: now },
        ];
      } else if (p === "/api/memories" && method === "GET")
        body = memories.filter((m) =>
          (m.title + m.content).includes(url.searchParams.get("q") || ""),
        );
      else if (p === "/api/memories" && method === "POST") {
        const m = {
          ...req.postDataJSON(),
          id: "new-memory",
          revision: 1,
          created_at: now,
          updated_at: now,
        };
        memories.push(m);
        body = m;
      } else if (p.endsWith("/pause")) {
        const n = nodes.find((n) => n.id === p.split("/")[3]);
        n.paused = req.postDataJSON().paused;
        body = n;
      } else if (p.startsWith("/api/nodes/") && method === "PATCH") {
        const n = nodes.find((n) => n.id === p.split("/")[3]);
        Object.assign(n, req.postDataJSON());
        body = n;
      } else if (p === "/api/nodes" && method === "POST") {
        body = {
          node: { id: "new-node", ...req.postDataJSON() },
          token: "TEST-ONLY-NOT-A-REAL-CREDENTIAL",
          owner_id: "test-owner",
          expires_at: now,
        };
      } else if (p === "/api/plan")
        body = {
          title: "자연어 테스트 계획",
          tasks: [
            {
              id: "one",
              node_id: "windows",
              tool: "system_status",
              args: {},
              depends_on: [],
              approval_level: "read-only",
            },
            {
              id: "two",
              node_id: "linux",
              tool: "system_status",
              args: {},
              depends_on: [],
              approval_level: "read-only",
            },
          ],
        };
      else if (p === "/api/runs" && method === "POST") {
        const r = {
          ...req.postDataJSON(),
          id: "new-run",
          status: "queued",
          created_at: now,
        };
        r.tasks.forEach((t) => (t.status = "queued"));
        runs.push(r);
        body = r;
      } else if (p.endsWith("/approval")) {
        const t = runs[0].tasks[1];
        t.status = req.postDataJSON().allow ? "queued" : "cancelled";
        runs[0].status = t.status;
        body = { ok: true };
      } else if (p.startsWith("/api/runs/"))
        body = runs.find((r) => r.id === p.split("/")[3]) || {};
      else body = { ok: true };
      return route.fulfill({
        contentType: "application/json",
        body: JSON.stringify(body),
      });
    });
    await page.reload();
    await page
      .getByRole("heading", { name: "모든 장비, 하나의 흐름." })
      .waitFor();
    fs.mkdirSync(path.join(root, "docs", "qa"), { recursive: true });
    if (!process.env.MARUBOT_UI_SKIP_SCREENSHOTS) await page.screenshot({
      path: path.join(root, "docs/qa/admin-overview-desktop.png"),
      fullPage: true,
    });
    await page
      .getByRole("button", { name: "이 작업 승인", exact: true })
      .click();
    await page.locator("#confirm button[value=ok]").click();
    await page.waitForTimeout(200);
    assert.equal(runs[0].tasks[1].status, "queued");
    await page
      .getByRole("button", { name: "기억 · Wiki", exact: true })
      .click();
    await page.getByRole("button", { name: "기억 추가", exact: true }).click();
    await page.locator("#memory-form [name=title]").fill("새 기억 UI 검증");
    await page
      .locator("#memory-form [name=content]")
      .fill("테스트용 기억 내용");
    await page.getByRole("button", { name: "기억 저장", exact: true }).click();
    await page
      .getByRole("heading", { name: "새 기억 UI 검증", exact: true })
      .waitFor();
    await page.getByRole("button", { name: "장비", exact: true }).click();
    await page
      .getByRole("button", { name: "설정 수정", exact: true })
      .first()
      .click();
    await page
      .locator("#edit-node-form [name=name]")
      .fill("Windows 사무실 · 수정 UI 검증");
    await page.getByRole("button", { name: "설정 저장", exact: true }).click();
    await page
      .getByRole("button", { name: "새 작업 중지", exact: true })
      .first()
      .click();
    await page
      .getByRole("button", { name: "새 작업 재개", exact: true })
      .first()
      .waitFor();
    assert.equal(nodes[0].paused, true);
    await page
      .getByRole("button", { name: "새 작업 재개", exact: true })
      .first()
      .click();
    await page.waitForFunction(
      () =>
        !document
          .querySelector('[data-action="node-pause"]')
          ?.textContent.includes("재개"),
    );
    await page.getByRole("button", { name: "장비 등록", exact: true }).click();
    await page.locator("#node-form [name=name]").fill("등록 UI 검증");
    await page
      .locator("#node-form [name=url]")
      .fill("https://100.100.0.3:8443");
    await page.locator("#node-form button[type=submit]").click();
    await page
      .getByRole("heading", { name: "장비 인증 자료", exact: true })
      .waitFor();
    await page.getByRole("button", { name: "닫기", exact: true }).click();
    assert.equal(await page.locator("#token").count(), 0);
    await page
      .getByRole("button", { name: "작업 만들기 ＋", exact: true })
      .click();
    await page.locator("#plan-form textarea").fill("두 장비 상태 확인");
    await page
      .getByRole("button", { name: "자연어로 계획 만들기", exact: true })
      .click();
    await page.waitForFunction(
      () => document.querySelectorAll("#steps .step").length === 2,
    );
    await page.getByRole("button", { name: "계획 제출", exact: true }).click();
    await page
      .getByRole("button", { name: "자연어 테스트 계획", exact: true })
      .waitFor();
    await page.getByRole("button", { name: "전체 현황", exact: true }).click();
    await page.setViewportSize({ width: 390, height: 844 });
    if (!process.env.MARUBOT_UI_SKIP_SCREENSHOTS) await page.screenshot({
      path: path.join(root, "docs/qa/admin-overview-mobile.png"),
      fullPage: true,
    });
    assert.equal(
      await page.evaluate(
        () => document.documentElement.scrollWidth > innerWidth,
      ),
      false,
      "모바일 가로 넘침",
    );
    assert.deepEqual(jsErrors, [], "브라우저 JS 오류");
    console.log(
      "PASS: 로그인 설정 상태, 미인증 API 차단, 현황, 개별 승인, 기억 저장, 장비 등록, 토큰 제거, 자연어 2단계 계획 제출, 모바일 폭, JS 오류 없음",
    );
  } finally {
    await browser?.close();
    child.kill();
  }
})().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
