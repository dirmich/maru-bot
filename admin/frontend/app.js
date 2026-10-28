import { version } from "./version.js";
import { forms } from "./forms.js";
export const $ = (s, root = document) => root.querySelector(s);
export const esc = (v) =>
  String(v ?? "").replace(
    /[&<>"']/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        c
      ],
  );
export const state = {
  me: null,
  view: "overview",
  overview: null,
  memories: [],
  events: [],
  query: "",
  stream: null,
};
export const date = (v) =>
  v && !v.startsWith("0001") ? new Date(v).toLocaleString("ko-KR") : "미보고";
const names = {
  queued: "대기",
  running: "실행 중",
  "need-approval": "승인 대기",
  succeeded: "완료",
  failed: "실패",
  cancelled: "취소",
  cancelling: "취소 요청 중",
  blocked: "선행 작업 실패",
  unknown: "실행 결과 확인 필요",
};
export const badge = (s) =>
  `<span class="badge ${esc(s)}">${esc(names[s] || s)}</span>`;
export const nodeName = (id) =>
  state.overview?.nodes.find((n) => n.id === id)?.name ||
  id ||
  "능력에 따라 자동 선택";
export const empty = (title, desc) =>
  `<div class="empty"><strong>${esc(title)}</strong><p>${esc(desc)}</p></div>`;
export async function api(path, options = {}) {
  const response = await fetch("/api" + path, {
    ...options,
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": state.me?.csrf || "",
      ...options.headers,
    },
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    if (response.status === 401 && state.me) {
      state.stream?.close();
      state.me = null;
      await login();
    }
    throw Error(data.error?.message || `요청 실패 (${response.status})`);
  }
  return data;
}
export function notify(message, error = false) {
  const el = $("#notice");
  if (el) {
    el.textContent = message;
    el.className = "notice" + (error ? " error" : "");
    el.hidden = false;
  }
}
export function editor(title, desc, html) {
  $("#editor-body").innerHTML =
    `<header class="editor-heading"><div><p class="eyebrow">내 작업 공간</p><h2>${esc(title)}</h2><p>${esc(desc)}</p></div><button id="close-editor" class="secondary" aria-label="닫기">×</button></header><div id="editor-error" class="notice error" hidden role="alert"></div>${html}`;
  $("#close-editor").onclick = () => $("#editor").close();
  if (!$("#editor").open) $("#editor").showModal();
}
export function editorError(error) {
  const el = $("#editor-error");
  el.textContent = error.message;
  el.hidden = false;
  el.scrollIntoView({ block: "nearest" });
}
export function confirm(title, message) {
  return new Promise((resolve) => {
    const d = $("#confirm");
    $("#confirm-title").textContent = title;
    $("#confirm-message").textContent = message;
    d.returnValue = "cancel";
    d.onclose = () => resolve(d.returnValue === "ok");
    d.showModal();
  });
}
async function login() {
  const auth = await api("/auth/status").catch(() => ({ configured: false }));
  $("#app").innerHTML =
    `<section class="login"><div class="wordmark"><b>m</b> marubot <small>control</small></div><p class="eyebrow">내 장비를 하나의 작업 공간으로</p><h1>전체 시스템을<br>한곳에서 관리하세요.</h1><p>장비별 병렬 작업과 실행 결과, 승인, 사용자 전용 기억을 확인합니다.</p>${auth.configured ? '<a class="primary button" href="/auth/google/login">Google로 로그인</a>' : '<div class="notice error">Google 로그인 설정이 필요합니다. backend의 GOOGLE_CLIENT_ID와 GOOGLE_CLIENT_SECRET을 설정하세요.</div>'}<small>사용자별 장비·작업·Wiki는 분리 보관됩니다. · v${esc(version)}</small></section>`;
}
function shell() {
  $("#app").innerHTML =
    `<a class="skip" href="#content">본문으로 이동</a><header class="topbar"><a href="/" class="wordmark"><b>m</b> marubot <small>control</small></a><div class="user"><span id="connection">연결 중</span><span>${esc(state.me.name || state.me.owner)}</span><button id="logout" class="text-button">로그아웃</button></div></header><div class="workspace"><div class="heading"><div><p class="eyebrow">시스템 관리 · v${esc(version)}</p><h1>모든 장비, 하나의 흐름.</h1><p class="muted">실행 상황과 필요한 조치를 한눈에 확인하세요.</p></div><div class="actions"><small id="updated"></small><button id="refresh" class="secondary">새로고침</button><button data-action="compose" class="primary">작업 만들기 ＋</button></div></div><div id="notice" class="notice" role="status" hidden></div><section id="metrics" class="metrics" aria-label="전체 현황"><div class="skeleton"></div></section><nav id="tabs" aria-label="관리 메뉴">${[
      ["overview", "전체 현황"],
      ["runs", "작업"],
      ["nodes", "장비"],
      ["memories", "기억 · Wiki"],
      ["audit", "활동 기록"],
    ]
      .map(
        ([id, label]) =>
          `<button data-view="${id}" class="${state.view === id ? "active" : ""}">${label}</button>`,
      )
      .join(
        "",
      )}</nav><section id="content" tabindex="-1" aria-live="polite">${empty("현황을 불러오고 있습니다", "잠시 기다려 주세요.")}</section><footer>MaruBot control · 사용자별로 분리된 작업 공간</footer></div><dialog id="editor"><div id="editor-body"></div></dialog><dialog id="confirm"><form method="dialog"><p class="eyebrow">작업 확인</p><h2 id="confirm-title"></h2><p id="confirm-message"></p><div class="actions"><button value="cancel" class="secondary">돌아가기</button><button value="ok" class="primary">확인</button></div></form></dialog>`;
  $("#logout").onclick = async () => {
    try {
      await api("/logout", { method: "POST", body: "{}" });
      state.stream?.close();
      state.me = null;
      await login();
    } catch (e) {
      notify(e.message, true);
    }
  };
  $("#refresh").onclick = refresh;
  $("#tabs").onclick = (e) => {
    const b = e.target.closest("[data-view]");
    if (b) {
      state.view = b.dataset.view;
      syncTabs();
      render();
    }
  };
  $("#app").onclick = onAction;
}
export function syncTabs() {
  $("#tabs")
    ?.querySelectorAll("button")
    .forEach((b) =>
      b.classList.toggle("active", b.dataset.view === state.view),
    );
}
let busy = false,
  timer;
export async function refresh() {
  if (busy || !state.me) return;
  busy = true;
  try {
    const [overview, memories, events] = await Promise.all([
      api("/overview"),
      api("/memories?q=" + encodeURIComponent(state.query)),
      api("/events"),
    ]);
    state.overview = overview;
    state.memories = memories;
    state.events = events;
    $("#metrics").innerHTML = [
      [
        "온라인 장비",
        `${overview.online}<small> / ${overview.nodes.length}</small>`,
        "연결된 사용자 소유 장비",
      ],
      ["실행 작업", overview.running, "병렬 수행 중인 개별 작업"],
      ["승인 대기", overview.pending, "사용자의 확인이 필요해요"],
      ["기억 · Wiki", overview.memory_count, "단기·장기·공유 기억"],
    ]
      .map(
        ([label, value, desc]) =>
          `<article><span>${label}</span><strong>${value}</strong><small>${desc}</small></article>`,
      )
      .join("");
    $("#updated").textContent =
      new Date(overview.updated_at).toLocaleTimeString("ko-KR") + " 갱신";
    render();
  } catch (e) {
    notify(e.message, true);
  } finally {
    busy = false;
  }
}
function connect() {
  state.stream?.close();
  const source = new EventSource("/api/events");
  state.stream = source;
  source.onopen = () => {
    if ($("#connection")) {
      $("#connection").textContent = "● 실시간 연결";
      $("#connection").className = "online";
    }
  };
  source.onerror = () => {
    if ($("#connection")) {
      $("#connection").textContent = "재연결 중 · 자동 갱신 유지";
      $("#connection").className = "muted";
    }
  };
  source.addEventListener("update", () => {
    clearTimeout(timer);
    timer = setTimeout(refresh, 250);
  });
}
function nodeTable() {
  const nodes = state.overview.nodes;
  return nodes.length
    ? `<div class="table-wrap"><table><thead><tr><th>장비</th><th>상태</th><th>플랫폼 · 버전</th><th>CPU / 메모리</th><th>허용 도구</th><th>마지막 보고</th><th>관리</th></tr></thead><tbody>${nodes.map((n) => `<tr><td><strong>${esc(n.name)}</strong><small>${esc(n.url)}</small></td><td><span class="dot ${n.online ? "online" : ""}"></span>${n.disabled ? "등록 폐기" : n.paused ? "새 작업 중지" : n.online ? "온라인" : "오프라인"}</td><td>${esc(n.platform || "미보고")} ${esc(n.arch)}<small>${esc(n.version || "—")}</small></td><td class="numeric">${Number(n.cpu || 0).toFixed(1)}% / ${Number(n.memory || 0).toFixed(1)}%</td><td><div class="tags">${(n.tools || []).map((t) => `<code>${esc(t)}</code>`).join("") || "—"}</div></td><td>${date(n.last_seen)}</td><td>${n.disabled ? "—" : `<button class="text-button" data-action="node-edit" data-id="${n.id}">설정 수정</button><button class="text-button" data-action="node-pause" data-id="${n.id}">${n.paused ? "새 작업 재개" : "새 작업 중지"}</button><button class="text-button" data-action="node-rotate" data-id="${n.id}">인증 갱신</button><button class="text-button danger" data-action="node-revoke" data-id="${n.id}">등록 폐기</button>`}</td></tr>`).join("")}</tbody></table></div>`
    : empty(
        "아직 연결된 장비가 없습니다",
        "장비를 등록하고 발급받은 설정으로 marubot mesh-worker를 실행하세요.",
      );
}
function runList(runs) {
  return runs.length
    ? `<div class="run-list">${[...runs]
        .sort(
          (a, b) =>
            Number(
              ["running", "need-approval", "queued", "unknown"].includes(
                b.status,
              ),
            ) -
              Number(
                ["running", "need-approval", "queued", "unknown"].includes(
                  a.status,
                ),
              ) || b.created_at.localeCompare(a.created_at),
        )
        .map(
          (r) =>
            `<article class="run-row"><div class="run-heading"><div><button class="run-title" data-action="run-detail" data-id="${r.id}">${esc(r.title)}</button><small>${date(r.created_at)} · ${r.tasks.length}개 작업</small></div>${badge(r.status)}</div><div class="dag">${r.tasks.map((t) => `<button class="dag-task ${esc(t.status)}" data-action="run-detail" data-id="${r.id}"><strong>${esc(t.id)} · ${esc(t.tool)}</strong><small>${esc(nodeName(t.node_id))}</small>${t.depends_on?.length ? `<small>선행: ${esc(t.depends_on.join(", "))}</small>` : ""}${badge(t.status)}</button>`).join("")}</div></article>`,
        )
        .join("")}</div>`
    : empty(
        "표시할 작업이 없습니다",
        "작업을 만들면 병렬 실행 상태와 결과를 확인할 수 있습니다.",
      );
}
export function memoryItems() {
  return state.memories.length
    ? state.memories
        .map(
          (m) =>
            `<article class="memory-item"><div class="memory-meta"><span class="badge">${{ short: "단기", long: "장기", wiki: "Wiki" }[m.kind]}</span><span>${m.visibility === "shared" ? "내 장비 공유" : "장비·세션 전용"}</span><span>r${m.revision}</span></div><h3>${esc(m.title)}</h3><p class="excerpt">${esc(m.content)}</p><small>${m.node_id ? esc(nodeName(m.node_id)) + " · " : ""}${m.expires_at ? "만료 " + date(m.expires_at) : "보존 기한 없음"}<br>갱신 ${date(m.updated_at)}</small><div class="memory-actions">${[["memory-edit", "보기 · 수정"], ["memory-history", "이력"], ...(m.kind !== "wiki" ? [["memory-promote", "Wiki 승격"]] : []), ["memory-delete", "삭제"]].map(([action, label]) => `<button class="text-button ${action === "memory-delete" ? "danger" : ""}" data-action="${action}" data-id="${m.id}">${label}</button>`).join("")}</div></article>`,
        )
        .join("")
    : empty(
        "기억이 없습니다",
        "장비 환경이나 검증한 결정을 기록하고 필요할 때 검색하세요.",
      );
}
function render() {
  if (state.view === "memories" && $("#memory-search")?.matches(":focus")) {
    if ($("#memory-items")) $("#memory-items").innerHTML = memoryItems();
    return;
  }
  if (!state.overview) return;
  const { runs } = state.overview;
  const section = (title, desc, action = "") =>
    `<div class="section-heading"><div><h2>${title}</h2><p>${desc}</p></div>${action}</div>`;
  let html = "";
  switch (state.view) {
    case "overview": {
      const pending = runs.flatMap((r) =>
        r.tasks
          .filter((t) => t.status === "need-approval")
          .map((t) => ({ r, t })),
      );
      html = `<div class="overview-grid"><section class="panel">${section("지금의 작업 흐름", "장비별 상태와 실행 순서를 살펴보세요.")} ${runList(runs.filter((r) => !["succeeded", "cancelled"].includes(r.status)).slice(0, 4))}</section><section class="panel">${section("확인이 필요한 작업", `${pending.length}개 작업이 승인 대기 중입니다.`)}${
        pending.length
          ? pending
              .slice(0, 6)
              .map(
                ({ r, t }) =>
                  `<article class="approval-item"><strong>${esc(t.instruction || t.id)}</strong><p>${esc(r.title)} · ${esc(nodeName(t.node_id))}</p><code>${esc(t.tool)}</code><pre>${esc(JSON.stringify(t.args || {}, null, 2))}</pre><div class="actions"><button class="secondary" data-action="deny" data-run="${r.id}" data-id="${esc(t.id)}">거부</button><button class="primary" data-action="approve" data-run="${r.id}" data-id="${esc(t.id)}">이 작업 승인</button></div></article>`,
              )
              .join("")
          : empty(
              "승인할 작업이 없습니다",
              "변경 작업이 접수되면 이곳에서 확인할 수 있습니다.",
            )
      }</section><section class="panel full">${section("장비 연결 상태", "heartbeat와 허용 도구를 확인하세요.", '<button class="secondary" data-action="node-new">장비 등록</button>')}${nodeTable()}</section></div>`;
      break;
    }
    case "runs":
      html =
        section(
          "작업과 실행 결과",
          "작업의 의존 관계와 실행 장비별 결과를 확인합니다.",
        ) + runList(runs);
      break;
    case "nodes":
      html =
        section(
          "내 장비",
          "마지막 heartbeat 이후 45초가 지나면 오프라인으로 표시합니다.",
          '<button class="primary" data-action="node-new">장비 등록</button>',
        ) + nodeTable();
      break;
    case "memories":
      html =
        section(
          "사용자 전용 기억 · Wiki",
          "공유 기억은 내 장비 사이에서만 공유됩니다. 단기 기억은 만료 후 조회되지 않습니다.",
          '<button class="primary" data-action="memory-new">기억 추가</button>',
        ) +
        `<label class="search">검색<input id="memory-search" type="search" value="${esc(state.query)}" placeholder="제목이나 본문에서 검색"></label><div class="memory-grid" id="memory-items">${memoryItems()}</div>`;
      break;
    case "audit":
      html =
        section(
          "활동 기록",
          "장비 등록, 승인, 작업 상태와 기억 변경을 기록합니다.",
        ) +
        (state.events.length
          ? `<ol class="timeline">${[...state.events]
              .reverse()
              .map(
                (e) =>
                  `<li><time>${date(e.time)}</time><div><strong>${esc(e.type)}</strong><p>${esc(e.message)}</p>${e.run_id ? `<button class="text-button" data-action="run-detail" data-id="${esc(e.run_id)}">작업 보기</button>` : ""}</div></li>`,
              )
              .join("")}</ol>`
          : empty(
              "아직 활동 기록이 없습니다",
              "장비를 등록하거나 작업을 만들면 이곳에 표시됩니다.",
            ));
      break;
  }
  $("#content").innerHTML = html;
  if ($("#memory-search"))
    $("#memory-search").oninput = (e) => {
      state.query = e.target.value;
      clearTimeout(timer);
      timer = setTimeout(async () => {
        try {
          state.memories = await api(
            "/memories?q=" + encodeURIComponent(state.query),
          );
          if ($("#memory-items")) $("#memory-items").innerHTML = memoryItems();
        } catch (e) {
          notify(e.message, true);
        }
      }, 300);
    };
}
export async function runDetail(id) {
  const r = await api("/runs/" + id);
  editor(
    r.title,
    `생성 ${date(r.created_at)} · 기한 ${date(r.deadline)}`,
    `${badge(r.status)}${r.status === "unknown" ? `<button class="secondary" data-action="run-reconcile" data-id="${r.id}">장비 실행 기록 다시 조회</button>` : ""}${r.tasks.map((t) => `<section class="task-detail"><div class="section-heading"><h3>${esc(t.id)} · ${esc(t.instruction || t.tool)}</h3>${badge(t.status)}</div><p>${esc(nodeName(t.node_id))} · <code>${esc(t.tool)}</code>${t.depends_on?.length ? " · 선행: " + esc(t.depends_on.join(", ")) : ""}</p><details><summary>실행 인자</summary><pre>${esc(JSON.stringify(t.args || {}, null, 2))}</pre></details>${t.output ? `<h4>결과</h4><pre>${esc(t.output)}</pre>` : ""}${t.error ? `<p class="danger">${esc(t.error)}</p>` : ""}${t.status === "need-approval" ? `<div class="actions"><button class="secondary" data-action="deny" data-run="${r.id}" data-id="${esc(t.id)}">거부</button><button class="primary" data-action="approve" data-run="${r.id}" data-id="${esc(t.id)}">이 작업 승인</button></div>` : ""}</section>`).join("")}<div class="actions">${["running", "queued", "need-approval"].includes(r.status) ? `<button class="secondary danger" data-action="run-cancel" data-id="${r.id}">작업 취소</button>` : ""}<button class="secondary" data-action="run-detail" data-id="${r.id}">결과 새로고침</button></div>`,
  );
}
async function onAction(event) {
  const b = event.target.closest("[data-action]");
  if (!b || b.disabled) return;
  b.disabled = true;
  try {
    await forms(b.dataset);
  } catch (e) {
    if ($("#editor")?.open) editorError(e);
    else notify(e.message, true);
  } finally {
    b.disabled = false;
  }
}
async function boot() {
  try {
    state.me = await api("/me");
    shell();
    await refresh();
    connect();
    setInterval(() => {
      if (
        state.me &&
        !document.hidden &&
        !$("#memory-search")?.matches(":focus")
      )
        refresh();
    }, 15000);
  } catch {
    await login();
  }
}
boot();
