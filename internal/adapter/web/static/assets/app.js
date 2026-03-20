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
      const msg = typeof data === "string" ? data : data?.error || res.statusText;
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

  const listProjects = $("#list-projects");
  const listHosts = $("#list-hosts");
  const selectProject = $("#select-project");
  const formHost = $("#form-host");
  const formProject = $("#form-project");
  const formToken = $("#form-token");
  const tokenChecks = $("#token-project-checks");
  const tokenResult = $("#token-result");
  const tokenValue = $("#token-value");
  const wrapPassword = $("#wrap-password");

  let projectsCache = [];

  async function loadProjects() {
    projectsCache = await api("GET", "api/projects");
    listProjects.innerHTML = "";
    selectProject.innerHTML = '<option value="">— выберите проект —</option>';
    tokenChecks.innerHTML = "";
    for (const p of projectsCache) {
      const li = document.createElement("li");
      li.innerHTML = `<span>${escapeHtml(p.name)}</span><span class="meta">${escapeHtml(p.id)}</span>`;
      const pick = document.createElement("button");
      pick.type = "button";
      pick.className = "pick";
      pick.textContent = "хосты";
      pick.addEventListener("click", () => {
        selectProject.value = p.id;
        selectProject.dispatchEvent(new Event("change"));
      });
      li.appendChild(pick);
      listProjects.appendChild(li);

      const opt = document.createElement("option");
      opt.value = p.id;
      opt.textContent = p.name;
      selectProject.appendChild(opt);

      const lab = document.createElement("label");
      lab.innerHTML = `<input type="checkbox" name="pid" value="${escapeAttr(p.id)}"> ${escapeHtml(p.name)}`;
      tokenChecks.appendChild(lab);
    }
    if (!projectsCache.length) {
      listProjects.innerHTML = '<li class="meta">пока нет проектов</li>';
    }
  }

  async function loadHosts(projectId) {
    if (!projectId) {
      listHosts.innerHTML = "";
      formHost.classList.add("hidden");
      return;
    }
    formHost.classList.remove("hidden");
    const hosts = await api("GET", `api/projects/${encodeURIComponent(projectId)}/hosts`);
    listHosts.innerHTML = "";
    if (!hosts.length) {
      listHosts.innerHTML = '<li class="meta">нет хостов</li>';
      return;
    }
    for (const h of hosts) {
      const li = document.createElement("li");
      li.innerHTML = `<span>${escapeHtml(h.name)} <span class="meta">${escapeHtml(h.username)}@${escapeHtml(h.address)}:${h.port}</span></span><span class="meta">${escapeHtml(h.auth_kind)}</span>`;
      listHosts.appendChild(li);
    }
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

  formProject.addEventListener("submit", async (e) => {
    e.preventDefault();
    const fd = new FormData(formProject);
    const name = fd.get("name").trim();
    if (!name) return;
    try {
      await api("POST", "api/projects", { name });
      formProject.reset();
      await loadProjects();
      toast("проект создан");
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
    const pid = selectProject.value;
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
      toast("хост добавлен");
    } catch (err) {
      toast(err.message, true);
    }
  });

  formToken.addEventListener("submit", async (e) => {
    e.preventDefault();
    const fd = new FormData(formToken);
    const label = String(fd.get("label") || "").trim();
    const project_ids = $$('input[name="pid"]:checked', tokenChecks).map((c) => c.value);
    try {
      const res = await api("POST", "api/tokens", { label, project_ids });
      tokenValue.textContent = res.token;
      tokenResult.classList.remove("hidden");
      tokenResult.focus();
      formToken.reset();
      toast("токен выпущен — сохраните");
    } catch (err) {
      toast(err.message, true);
    }
  });

  $("#btn-copy-token").addEventListener("click", async () => {
    const t = tokenValue.textContent;
    try {
      await navigator.clipboard.writeText(t);
      toast("скопировано");
    } catch {
      toast("не удалось скопировать", true);
    }
  });

  $("#btn-refresh-projects").addEventListener("click", () => {
    loadProjects().then(() => toast("список обновлён")).catch((err) => toast(err.message, true));
  });

  loadProjects().catch((err) => toast(err.message, true));
})();
