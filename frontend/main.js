let secondsLeft = 0;
let totalSeconds = 0;
let timerInterval = null;

function getThemeFromStylesheetHref(href) {
    const match = href.match(/themes\/(theme\d+)\.css$/);
    return match ? match[1] : "theme1";
}

// ── Init ──────────────────────────────────────────
(async function init() {
    const themeLink = document.getElementById("theme-stylesheet");
    const themeSelect = document.getElementById("theme-select");
    let theme = getThemeFromStylesheetHref(themeLink.getAttribute("href") || "");

    if (window.go && window.go.main && window.go.main.App) {
        const status = await window.go.main.App.GetStatus();
        document.getElementById("status-label").textContent = `Ready. Default bout: ${status.boutMinutes} min.`;
        document.getElementById("outcome-input").placeholder = `What will you PRODUCE in the next ${status.boutMinutes} minutes?`;
        document.getElementById("bout-minutes-input").value = status.boutMinutes;
        if (status.theme) {
            theme = status.theme;
        }

        if (status.hasOutcome) {
            const prefix = status.isShifted ? "Shifted" : "Ready";
            document.getElementById("outcome-display").style.display = "block";
            document.getElementById("outcome-display").textContent = `${prefix}: ${status.outcome}`;
            document.getElementById("status-label").style.display = "none";
        }
    }

    themeSelect.value = theme;
    themeLink.href = `themes/${theme}.css`;
})();
// ── Timer ─────────────────────────────────────────
async function startBout() {
    const input = document.getElementById("outcome-input");
    const outcome = input.value.trim();
    if (!outcome) return;

    await window.go.main.App.SetOutcome(outcome);
    await window.go.main.App.StartBout();

    const status = await window.go.main.App.GetStatus();
    secondsLeft = status.boutMinutes * 60;
    totalSeconds = secondsLeft;

    document.getElementById("idle-area").style.display = "none";
    document.getElementById("running-area").style.display = "block";
    document.getElementById("outcome-display").style.display = "block";
    document.getElementById("outcome-display").textContent = outcome;
    document.getElementById("status-label").style.display = "none";

    updateTimer();
    timerInterval = setInterval(() => {
        secondsLeft--;
        updateTimer();
        if (secondsLeft <= 0) {
            clearInterval(timerInterval);
            document.getElementById("running-area").style.display = "none";
            document.getElementById("review-text").textContent = `Did you produce: "${outcome}"?`;
            document.getElementById("review-area").style.display = "block";
        }
    }, 1000);
}

async function cancelBout() {
    clearInterval(timerInterval);
    await window.go.main.App.CancelBout();
    resetUI();
}

async function review(result) {
    await window.go.main.App.Review(result);
    if (result === "shift") {
        const status = await window.go.main.App.GetStatus();
        document.getElementById("outcome-display").style.display = "block";
        document.getElementById("outcome-display").textContent = `Shifted: ${status.outcome}`;
        document.getElementById("status-label").style.display = "none";
    } else {
        document.getElementById("outcome-display").style.display = "none";
        document.getElementById("status-label").style.display = "block";
    }
    resetUI();
}

function updateTimer() {
    const mins = Math.floor(secondsLeft / 60);
    const secs = secondsLeft % 60;
    document.getElementById("timer").textContent = `${mins}:${secs.toString().padStart(2, "0")}`;
}

function resetUI() {
    clearInterval(timerInterval);
    document.getElementById("idle-area").style.display = "block";
    document.getElementById("running-area").style.display = "none";
    document.getElementById("review-area").style.display = "none";
    document.getElementById("outcome-input").value = "";
    secondsLeft = 0;
    totalSeconds = 0;
}

// ── Settings ──────────────────────────────────────
async function toggleSettings() {
    const panel = document.getElementById("settings-panel");
    panel.style.display = panel.style.display === "none" ? "block" : "none";
    document.getElementById("audio-panel").style.display = "none";
    document.getElementById("log-panel").style.display = "none";
}

async function saveSettings() {
    const mins = parseInt(document.getElementById("bout-minutes-input").value);
    if (mins < 1 || mins > 180) return;
    await window.go.main.App.SetBoutMinutes(mins);
    document.getElementById("status-label").textContent = `Ready. Default bout: ${mins} min.`;
    document.getElementById("outcome-input").placeholder = `What will you PRODUCE in the next ${mins} minutes?`;
    document.getElementById("settings-panel").style.display = "none";
}

// ── Audio ─────────────────────────────────────────
async function toggleAudio() {
    const panel = document.getElementById("audio-panel");
    panel.style.display = panel.style.display === "none" ? "block" : "none";
    document.getElementById("settings-panel").style.display = "none";
    document.getElementById("log-panel").style.display = "none";
}

async function toggleChimes() { /* TODO: wire to Go */ }
async function toggleMusic() { /* TODO: wire to Go */ }

// ── Log ───────────────────────────────────────────
async function toggleLog() {
    const panel = document.getElementById("log-panel");
    const idle = document.getElementById("idle-area");
    const running = document.getElementById("running-area");
    const review = document.getElementById("review-area");
    const settings = document.getElementById("settings-panel");
    const audio = document.getElementById("audio-panel");

    if (panel.style.display === "flex") {
        panel.style.display = "none";
        if (idle) idle.style.display = "block";
        return;
    }

    // Hide other panels
    if (idle) idle.style.display = "none";
    if (running) running.style.display = "none";
    if (review) review.style.display = "none";
    if (settings) settings.style.display = "none";
    if (audio) audio.style.display = "none";

    // Fetch grouped log
    const days = await window.go.main.App.GetLog();
    const stats = await window.go.main.App.GetStats();

    let html = `<div class="log-stats">${stats.totalSessions} sessions · ${stats.completed} done · ${stats.shifted} shifted · ${stats.discarded} discarded</div>`;

    if (days.length === 0) {
        html += `<div style="padding: 20px; text-align: center; color: #999;">No sessions yet. Complete a bout to see logs.</div>`;
    } else {
        days.forEach(day => {
            const displayDate = new Date(day.date + "T00:00:00").toLocaleDateString("en-US", {
                weekday: "short",
                month: "short",
                day: "numeric",
            });
            const hours = (day.totalMinutes / 60).toFixed(1);
            html += `<div class="log-date">${displayDate} · ${hours}h worked</div>`;
            html += `<table>`;
            html += `<tr><th></th><th>Outcome</th><th>Min</th><th>Start</th><th>End</th></tr>`;
            day.sessions.forEach(s => {
                const icon = { completed: "✓", shifted: "↪", discarded: "✗" }[s.status] || "?";
                const startTime = s.createdAt ? s.createdAt.slice(11, 16) : "";
                const endTime = s.completedAt ? s.completedAt.slice(11, 16) : "";
                html += `<tr>
                    <td>${icon}</td>
                    <td>${s.plannedOutcome}</td>
                    <td>${s.boutMinutes}</td>
                    <td>${startTime}</td>
                    <td>${endTime}</td>
                </tr>`;
            });
            html += `</table>`;
        });
    }

    document.getElementById("log-content").innerHTML = html;
    panel.style.display = "flex";
}
function changeTheme(theme) {
    document.getElementById("theme-stylesheet").href = `themes/${theme}.css`;
    if (window.go && window.go.main && window.go.main.App && typeof window.go.main.App.SetTheme === "function") {
        window.go.main.App.SetTheme(theme);
    }
}