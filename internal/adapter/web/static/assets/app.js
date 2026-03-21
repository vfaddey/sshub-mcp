(function () {
  const $ = (sel, el = document) => el.querySelector(sel);
  const $$ = (sel, el = document) => [...el.querySelectorAll(sel)];

  function apiURL(path) {
    return new URL(path, window.location.href).href;
  }

  async function api(method, path, body) {
    const opt = { method, headers: {} };
    if (body !== undefined) {
      opt.headers["Content-Type"] = "application/json";
      opt.body = JSON.stringify(body);
    }
    const res = await fetch(apiURL(path), opt);
    const text = await res.text();
    let data;
    try {
      data = text ? JSON.parse(text) : null;
    } catch {
      data = text;
    }
    if (!res.ok) {
      const msg =
        typeof data === "string" ? data : data?.error || res.statusText;
      throw new Error(msg || "request failed");
    }
    return data;
  }

  function toast(msg, err) {
    const t = $("#toast");
    t.textContent = msg;
    t.classList.toggle("err", !!err);
    t.classList.remove("hidden");
    clearTimeout(t._hide);
    t._hide = setTimeout(() => t.classList.add("hidden"), 4200);
  }

  function toIntId(v) {
    const s = String(v ?? "").trim();
    if (!s) return null;
    const n = Number(s);
    if (!Number.isInteger(n) || n <= 0) return null;
    return n;
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function escapeAttr(s) {
    return escapeHtml(s).replace(/'/g, "&#39;");
  }

  async function withConfirm(msg, fn) {
    if (!confirm(msg)) return;
    return await fn();
  }

  const listProjects = $("#list-projects");
  const listHosts = $("#list-hosts");
  const listTokens = $("#list-tokens");
  const selectProject = $("#select-project");
  const formHost = $("#form-host");
  const formProject = $("#form-project");
  const formToken = $("#form-token");
  const tokenChecks = $("#token-project-checks");
  const tokenResult = $("#token-result");
  const tokenValue = $("#token-value");
  const wrapPassword = $("#wrap-password");

  let projectsCache = [];

  async function loadTokens() {
    if (!listTokens) return;
    const tokens = await api("GET", "api/tokens");
    listTokens.innerHTML = "";
    if (!tokens.length) {
      listTokens.innerHTML = '<li class="meta">no tokens</li>';
      return;
    }
    for (const t of tokens) {
      const li = document.createElement("li");
      const left = document.createElement("span");
      left.innerHTML = `${escapeHtml(t.label)} <span class="meta">#${escapeHtml(t.id)}</span>`;

      const actions = document.createElement("span");
      actions.className = "actions";

      const del = document.createElement("button");
      del.type = "button";
      del.className = "btn danger icon";
      del.setAttribute("aria-label", "Delete token");
      del.setAttribute("title", "Delete");
      del.innerHTML = `<svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M9 3h6l1 2h4v2H4V5h4l1-2zm1 7h2v9h-2v-9zm4 0h2v9h-2v-9zM7 10h2v9H7v-9z"/></svg>`;
      del.addEventListener("click", async () => {
        try {
          await withConfirm(`Delete token #${t.id}?`, async () => {
            await api("DELETE", `api/tokens/${encodeURIComponent(t.id)}`);
          });
          await loadTokens();
          toast("token deleted");
        } catch (err) {
          toast(err.message, true);
        }
      });

      actions.appendChild(del);
      li.appendChild(left);
      li.appendChild(actions);
      listTokens.appendChild(li);
    }
  }

  async function loadProjects() {
    projectsCache = await api("GET", "api/projects");
    listProjects.innerHTML = "";
    selectProject.innerHTML = '<option value="">— select a project —</option>';
    tokenChecks.innerHTML = "";

    for (const p of projectsCache) {
      const li = document.createElement("li");

      const left = document.createElement("span");
      left.innerHTML = `${escapeHtml(p.name)} <span class="meta">${escapeHtml(p.id)}</span>`;

      const actions = document.createElement("span");
      actions.className = "actions";

      const pick = document.createElement("button");
      pick.type = "button";
      pick.className = "pick";
      pick.textContent = "hosts";
      pick.addEventListener("click", () => {
        selectProject.value = String(p.id);
        selectProject.dispatchEvent(new Event("change"));
      });

      const del = document.createElement("button");
      del.type = "button";
      del.className = "btn danger icon";
      del.setAttribute("aria-label", `Delete project "${p.name}"`);
      del.setAttribute("title", "Delete");
      del.innerHTML = `<svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M9 3h6l1 2h4v2H4V5h4l1-2zm1 7h2v9h-2v-9zm4 0h2v9h-2v-9zM7 10h2v9H7v-9z"/></svg>`;
      del.addEventListener("click", async () => {
        try {
          await withConfirm(
            `Delete project "${p.name}"? This will also delete its hosts.`,
            async () => {
              await api("DELETE", `api/projects/${encodeURIComponent(p.id)}`);
            },
          );
          if (String(selectProject.value) === String(p.id)) {
            selectProject.value = "";
            listHosts.innerHTML = "";
            formHost.classList.add("hidden");
          }
          await loadProjects();
          await loadTokens();
          toast("project deleted");
        } catch (err) {
          toast(err.message, true);
        }
      });

      actions.appendChild(pick);
      actions.appendChild(del);

      li.appendChild(left);
      li.appendChild(actions);
      listProjects.appendChild(li);

      const opt = document.createElement("option");
      opt.value = String(p.id);
      opt.textContent = p.name;
      selectProject.appendChild(opt);

      const lab = document.createElement("label");
      lab.innerHTML = `<input type="checkbox" name="pid" value="${escapeAttr(p.id)}"> ${escapeHtml(p.name)}`;
      tokenChecks.appendChild(lab);
    }

    if (!projectsCache.length) {
      listProjects.innerHTML = '<li class="meta">no projects yet</li>';
    }
  }

  async function loadHosts(projectIdRaw) {
    const projectId = toIntId(projectIdRaw);
    if (!projectId) {
      listHosts.innerHTML = "";
      formHost.classList.add("hidden");
      return;
    }

    formHost.classList.remove("hidden");
    const hosts = await api(
      "GET",
      `api/projects/${encodeURIComponent(projectId)}/hosts`,
    );

    listHosts.innerHTML = "";
    if (!hosts.length) {
      listHosts.innerHTML = '<li class="meta">no hosts</li>';
      return;
    }

    for (const h of hosts) {
      const li = document.createElement("li");

      const left = document.createElement("span");
      left.innerHTML = `${escapeHtml(h.name)} <span class="meta">${escapeHtml(h.username)}@${escapeHtml(h.address)}:${h.port}</span>`;

      const actions = document.createElement("span");
      actions.className = "actions";

      const meta = document.createElement("span");
      meta.className = "meta";
      meta.textContent = h.auth_kind;

      const del = document.createElement("button");
      del.type = "button";
      del.className = "btn danger icon";
      del.setAttribute("aria-label", `Delete host "${h.name}"`);
      del.setAttribute("title", "Delete");
      del.innerHTML = `<svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M9 3h6l1 2h4v2H4V5h4l1-2zm1 7h2v9h-2v-9zm4 0h2v9h-2v-9zM7 10h2v9H7v-9z"/></svg>`;
      del.addEventListener("click", async () => {
        try {
          await withConfirm(`Delete host "${h.name}"?`, async () => {
            await api(
              "DELETE",
              `api/projects/${encodeURIComponent(projectId)}/hosts/${encodeURIComponent(h.id)}`,
            );
          });
          await loadHosts(projectId);
          toast("host deleted");
        } catch (err) {
          toast(err.message, true);
        }
      });

      actions.appendChild(meta);
      actions.appendChild(del);

      li.appendChild(left);
      li.appendChild(actions);
      listHosts.appendChild(li);
    }
  }

  formProject.addEventListener("submit", async (e) => {
    e.preventDefault();
    const fd = new FormData(formProject);
    const name = String(fd.get("name") || "").trim();
    if (!name) return;
    try {
      await api("POST", "api/projects", { name });
      formProject.reset();
      await loadProjects();
      await loadTokens();
      toast("project created");
    } catch (err) {
      toast(err.message, true);
    }
  });

  selectProject.addEventListener("change", () => {
    loadHosts(selectProject.value).catch((err) => toast(err.message, true));
  });

  formHost.querySelector("[name=auth_kind]").addEventListener("change", (e) => {
    wrapPassword.classList.toggle("hidden", e.target.value !== "password");
  });

  formHost.addEventListener("submit", async (e) => {
    e.preventDefault();
    const pid = toIntId(selectProject.value);
    if (!pid) return;

    const fd = new FormData(formHost);
    const port = parseInt(String(fd.get("port") || "22"), 10);
    const body = {
      name: String(fd.get("name") || "").trim(),
      address: String(fd.get("address") || "").trim(),
      port: Number.isFinite(port) ? port : 22,
      username: String(fd.get("username") || "").trim(),
      auth_kind: String(fd.get("auth_kind") || "none"),
      password: String(fd.get("password") || ""),
    };

    try {
      await api("POST", `api/projects/${encodeURIComponent(pid)}/hosts`, body);
      formHost.reset();
      formHost.querySelector("[name=port]").value = "22";
      wrapPassword.classList.add("hidden");
      await loadHosts(pid);
      toast("host added");
    } catch (err) {
      toast(err.message, true);
    }
  });

  formToken.addEventListener("submit", async (e) => {
    e.preventDefault();
    const fd = new FormData(formToken);
    const label = String(fd.get("label") || "").trim();
    const project_ids = $$('input[name="pid"]:checked', tokenChecks)
      .map((c) => toIntId(c.value))
      .filter((v) => v !== null);

    try {
      const res = await api("POST", "api/tokens", { label, project_ids });
      tokenValue.textContent = res.token;
      tokenResult.classList.remove("hidden");
      tokenResult.focus();
      formToken.reset();
      await loadTokens();
      toast("token issued — save it now");
    } catch (err) {
      toast(err.message, true);
    }
  });

  $("#btn-copy-token").addEventListener("click", async () => {
    const t = tokenValue.textContent;
    try {
      await navigator.clipboard.writeText(t);
      toast("copied");
    } catch {
      toast("failed to copy", true);
    }
  });

  $("#btn-refresh-projects").addEventListener("click", () => {
    loadProjects()
      .then(() => toast("refreshed"))
      .catch((err) => toast(err.message, true));
  });

  const refreshTokensBtn = $("#btn-refresh-tokens");
  if (refreshTokensBtn) {
    refreshTokensBtn.addEventListener("click", () => {
      loadTokens()
        .then(() => toast("refreshed"))
        .catch((err) => toast(err.message, true));
    });
  }

  Promise.all([loadProjects(), loadTokens()]).catch((err) =>
    toast(err.message, true),
  );
})();
