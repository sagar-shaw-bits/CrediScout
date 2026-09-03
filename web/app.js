const chat = document.getElementById("chat");
const form = document.getElementById("form");
const input = document.getElementById("q");
const sendBtn = document.getElementById("send");

function addMessage(cls, text) {
  const el = document.createElement("div");
  el.className = `msg ${cls}`;
  el.textContent = text;
  chat.appendChild(el);
  chat.scrollTop = chat.scrollHeight;
  return el;
}

function renderSources(el, sources) {
  if (!sources || sources.length === 0) return;
  const box = document.createElement("div");
  box.className = "sources";
  box.innerHTML = "<strong>Sources:</strong>";
  sources.forEach((s) => {
    const a = document.createElement("a");
    a.href = s.url;
    a.target = "_blank";
    a.rel = "noopener noreferrer";
    const score = typeof s.score === "number" ? ` (score ${s.score.toFixed(2)})` : "";
    a.textContent = `${s.title || s.url}${score}`;
    box.appendChild(a);
  });
  el.appendChild(box);
  chat.scrollTop = chat.scrollHeight;
}

async function ask(question) {
  addMessage("user", question);
  const botEl = addMessage("bot", "…");
  sendBtn.disabled = true;

  try {
    const resp = await fetch("/api/ask", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ question }),
    });
    if (!resp.ok) {
      const err = await resp.json().catch(() => ({}));
      throw new Error(err.error || `Request failed (${resp.status})`);
    }
    const data = await resp.json();
    botEl.textContent = data.answer || "";
    renderSources(botEl, data.sources);
  } catch (e) {
    botEl.textContent = `⚠️ ${e.message}`;
  } finally {
    sendBtn.disabled = false;
    input.focus();
  }
}

form.addEventListener("submit", (e) => {
  e.preventDefault();
  const q = input.value.trim();
  if (!q) return;
  input.value = "";
  ask(q);
});

document.querySelectorAll(".chip").forEach((chip) => {
  chip.addEventListener("click", () => {
    input.value = chip.textContent;
    input.focus();
  });
});
