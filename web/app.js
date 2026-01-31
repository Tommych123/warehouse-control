let token = localStorage.getItem("token") || "";
let me = null;
let selectedItemId = null;

function qs(sel) { return document.querySelector(sel); }
function out(obj) { qs("#out").textContent = typeof obj === "string" ? obj : JSON.stringify(obj, null, 2); }
function setWhoami() {
  const el = qs("#whoami");
  if (!me) {
    el.textContent = "Не авторизован";
    return;
  }
  el.textContent = `${me.username} (${me.role})`;
}
function roleRank(role) {
  if (role === "admin") return 3;
  if (role === "manager") return 2;
  if (role === "viewer") return 1;
  return 0;
}
function canWrite() { return me && roleRank(me.role) >= 2; }
function canDelete() { return me && me.role === "admin"; }

async function api(path, opts = {}) {
  const headers = opts.headers || {};
  if (token) headers["Authorization"] = "Bearer " + token;
  return fetch(path, { ...opts, headers });
}

async function login() {
  const username = qs("#username").value.trim();
  const role = qs("#role").value;

  const res = await fetch("/api/auth/login", {
    method: "POST",
    headers: {"Content-Type":"application/json"},
    body: JSON.stringify({username, role})
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) { out({status: res.status, ...data}); return; }

  token = data.token || "";
  localStorage.setItem("token", token);
  await loadMe();
  await loadItems();
}

function logout() {
  token = "";
  me = null;
  localStorage.removeItem("token");
  setWhoami();
  setPermissionsUI();
  clearTables();
  out("logged out");
}

async function loadMe() {
  if (!token) { me = null; return; }
  const res = await api("/api/me");
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    me = null;
    out({status: res.status, ...data});
    return;
  }
  me = data;
  setWhoami();
  setPermissionsUI();
}

function setPermissionsUI() {
  const hint = qs("#permHint");
  if (!me) {
    hint.textContent = "Нужно войти чтобы работать.";
    qs("#btnLoadHistory").disabled = true;
    qs("#btnExportCsv").disabled = true;
    qs("#btnSave").disabled = true;
    return;
  }

  qs("#btnSave").disabled = !canWrite();
  hint.textContent = canWrite()
    ? (canDelete() ? "Роль: можно создавать/редактировать/удалять." : "Роль: можно создавать/редактировать. Удаление запрещено.")
    : "Роль: только просмотр.";

  qs("#btnLoadHistory").disabled = selectedItemId == null;
  qs("#btnExportCsv").disabled = selectedItemId == null;
}

function clearTables() {
  qs("#itemsTable tbody").innerHTML = "";
  qs("#historyTable tbody").innerHTML = "";
  qs("#histItemId").textContent = "—";
  selectedItemId = null;
}

async function loadItems() {
  if (!me) return;
  const search = qs("#search").value.trim();
  const url = search ? `/api/items?search=${encodeURIComponent(search)}` : "/api/items";

  const res = await api(url);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) { out({status: res.status, ...data}); return; }

  renderItems(data);
  out({items: data.length});
}

function renderItems(items) {
  const tbody = qs("#itemsTable tbody");
  tbody.innerHTML = "";

  for (const it of items) {
    const tr = document.createElement("tr");

    const actions = document.createElement("div");
    actions.className = "row";

    const btnHist = document.createElement("button");
    btnHist.className = "secondary";
    btnHist.textContent = "History";
    btnHist.onclick = () => selectItemForHistory(it.id);

    const btnEdit = document.createElement("button");
    btnEdit.className = "secondary";
    btnEdit.textContent = "Edit";
    btnEdit.disabled = !canWrite();
    btnEdit.onclick = () => fillForm(it);

    const btnDel = document.createElement("button");
    btnDel.className = "secondary";
    btnDel.textContent = "Delete";
    btnDel.disabled = !canDelete();
    btnDel.onclick = () => deleteItem(it.id);

    actions.appendChild(btnHist);
    actions.appendChild(btnEdit);
    actions.appendChild(btnDel);

    tr.innerHTML = `
      <td>${it.id}</td>
      <td>${escapeHtml(it.sku)}</td>
      <td>${escapeHtml(it.name)}</td>
      <td>${it.qty}</td>
      <td>${it.location ? escapeHtml(it.location) : ""}</td>
      <td></td>
    `;
    tr.children[5].appendChild(actions);
    tbody.appendChild(tr);
  }
}

function fillForm(it) {
  qs("#itemId").value = it.id;
  qs("#sku").value = it.sku;
  qs("#name").value = it.name;
  qs("#qty").value = it.qty;
  qs("#location").value = it.location || "";
  qs("#formTitle").textContent = "Редактировать";
}

function resetForm() {
  qs("#itemId").value = "";
  qs("#sku").value = "";
  qs("#name").value = "";
  qs("#qty").value = 0;
  qs("#location").value = "";
  qs("#formTitle").textContent = "Добавить / Редактировать";
}

async function saveItem(e) {
  e.preventDefault();
  if (!canWrite()) { out("no permission"); return; }

  const id = qs("#itemId").value.trim();
  const payload = {
    sku: qs("#sku").value.trim(),
    name: qs("#name").value.trim(),
    qty: Number(qs("#qty").value),
    location: qs("#location").value.trim() || null
  };

  const isUpdate = id !== "";
  const res = await api(isUpdate ? `/api/items/${id}` : "/api/items", {
    method: isUpdate ? "PUT" : "POST",
    headers: {"Content-Type":"application/json"},
    body: JSON.stringify(payload)
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) { out({status: res.status, ...data}); return; }

  out(data);
  resetForm();
  await loadItems();
}

async function deleteItem(id) {
  if (!canDelete()) return;
  if (!confirm(`Delete item #${id}?`)) return;

  const res = await api(`/api/items/${id}`, { method: "DELETE" });
  if (res.status === 204) {
    out({deleted: id});
    await loadItems();
    if (selectedItemId === id) {
      qs("#historyTable tbody").innerHTML = "";
      qs("#histItemId").textContent = "—";
      selectedItemId = null;
      setPermissionsUI();
    }
    return;
  }
  const data = await res.json().catch(() => ({}));
  out({status: res.status, ...data});
}

async function selectItemForHistory(id) {
  selectedItemId = id;
  qs("#histItemId").textContent = String(id);
  qs("#btnLoadHistory").disabled = false;
  qs("#btnExportCsv").disabled = false;
  setPermissionsUI();
  await loadHistory();
}

function historyQuery() {
  const params = new URLSearchParams();
  const from = qs("#fFrom").value.trim();
  const to = qs("#fTo").value.trim();
  const user = qs("#fUser").value.trim();
  const action = qs("#fAction").value.trim();
  const diff = qs("#fDiff").value;

  if (from) params.set("from", from);
  if (to) params.set("to", to);
  if (user) params.set("user", user);
  if (action) params.set("action", action);
  if (diff === "1") params.set("includeChanges", "1");

  const qsStr = params.toString();
  return qsStr ? "?" + qsStr : "";
}

async function loadHistory() {
  if (!selectedItemId) return;

  const res = await api(`/api/items/${selectedItemId}/history${historyQuery()}`);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) { out({status: res.status, ...data}); return; }

  renderHistory(data);
  out({history: data.length});
}

function renderHistory(entries) {
  const tbody = qs("#historyTable tbody");
  tbody.innerHTML = "";

  for (const e of entries) {
    const tr = document.createElement("tr");

    const diffText = e.changes ? JSON.stringify(e.changes, null, 2) : "";
    tr.innerHTML = `
      <td>${escapeHtml(new Date(e.changed_at).toISOString())}</td>
      <td>${escapeHtml(e.action)}</td>
      <td>${escapeHtml(e.actor || "")}</td>
      <td>${escapeHtml(e.actor_role || "")}</td>
      <td><pre class="mini">${escapeHtml(diffText)}</pre></td>
    `;
    tbody.appendChild(tr);
  }
}

async function exportCSV() {
  if (!selectedItemId) return;
  const url = `/api/items/${selectedItemId}/history.csv${historyQuery()}`;

  const res = await api(url);
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    out({status: res.status, ...data});
    return;
  }

  const blob = await res.blob();
  const a = document.createElement("a");
  a.href = URL.createObjectURL(blob);
  a.download = `history_item_${selectedItemId}.csv`;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(a.href);
}

function escapeHtml(s) {
  return String(s)
    .replaceAll("&","&amp;")
    .replaceAll("<","&lt;")
    .replaceAll(">","&gt;")
    .replaceAll('"',"&quot;")
    .replaceAll("'","&#039;");
}

qs("#btnLogin").addEventListener("click", login);
qs("#btnLogout").addEventListener("click", logout);
qs("#btnRefresh").addEventListener("click", loadItems);
qs("#itemForm").addEventListener("submit", saveItem);
qs("#btnReset").addEventListener("click", resetForm);
qs("#btnLoadHistory").addEventListener("click", loadHistory);
qs("#btnExportCsv").addEventListener("click", exportCSV);

// on load
(async function init() {
  await loadMe();
  if (me) await loadItems();
  setPermissionsUI();
})();
