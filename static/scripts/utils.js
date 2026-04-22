// =========================================================
// UTILIDADES COMPARTILHADAS
// =========================================================

// =========================================================
// AUTH & USER
// =========================================================

function authHeaders() {
    return {
        "Content-Type": "application/json",
        "Authorization": "Bearer " + localStorage.getItem("auth_token")
    };
}

function getUser() {
    try {
        return JSON.parse(localStorage.getItem("user_payload")) || {};
    } catch {
        return {};
    }
}

// =========================================================
// FORMATAÇÃO
// =========================================================

function formatDate(dateStr) {
    return new Date(dateStr).toLocaleString("pt-BR");
}

function escapeHtml(str) {
    return String(str)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;");
}

// =========================================================
// MENSAGENS
// =========================================================

function showMessage(containerId, text, type = "error", duration = 3000) {
    const el = document.getElementById(containerId);
    if (!el) return;

    el.className = `message ${type}`;
    el.textContent = text;

    if (duration > 0) {
        setTimeout(() => {
            el.textContent = "";
            el.className = "message";
        }, duration);
    }
}

// =========================================================
// MODAL GENÉRICA
// =========================================================

function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.remove("hidden");
        document.body.style.overflow = "hidden";
    }
}

function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add("hidden");
        document.body.style.overflow = "auto";
    }
}

function setupModalToggle(btnId, modalId, cancelBtnId) {
    const btn = document.getElementById(btnId);
    const modal = document.getElementById(modalId);
    const cancelBtn = document.getElementById(cancelBtnId);

    if (!btn || !modal) return;

    btn.onclick = () => openModal(modalId);

    if (cancelBtn) {
        cancelBtn.onclick = () => closeModal(modalId);
    }

    modal.onclick = (e) => {
        if (e.target === modal) closeModal(modalId);
    };

    document.addEventListener("keydown", (e) => {
        if (e.key === "Escape") closeModal(modalId);
    });
}

// =========================================================
// PAGINAÇÃO GENÉRICA
// =========================================================

function renderPagination(containerId, currentPage, hasNext, onPageChange) {
    const container = document.getElementById(containerId);
    if (!container) return;

    container.innerHTML = `
        <button class="page-btn" ${currentPage === 1 ? "disabled" : ""} id="prevPage">
            ◀
        </button>

        <span class="page-number">${currentPage}</span>

        <button class="page-btn" ${!hasNext ? "disabled" : ""} id="nextPage">
            ▶
        </button>
    `;

    const prev = document.getElementById("prevPage");
    const next = document.getElementById("nextPage");

    if (prev) prev.onclick = () => onPageChange(currentPage - 1);
    if (next) next.onclick = () => onPageChange(currentPage + 1);
}

// =========================================================
// FETCH HELPERS
// =========================================================

async function fetchAPI(url, method = "GET", body = null) {
    try {
        const options = {
            method,
            headers: authHeaders()
        };

        if (body) {
            options.body = JSON.stringify(body);
        }

        const res = await fetch(url, options);

        if (!res.ok) {
            throw new Error(`API Error: ${res.status}`);
        }

        return await res.json();
    } catch (err) {
        console.error(`Erro em ${method} ${url}:`, err);
        throw err;
    }
}

// =========================================================
// SCRIPT LOADER COM CACHE-BUSTING
// =========================================================

function loadScript(src, moduleId) {
    return new Promise((resolve, reject) => {
        document.querySelectorAll("script[data-module-script]").forEach(el => el.remove());

        const script = document.createElement("script");
        script.src = src + "?v=" + Date.now();
        script.setAttribute("data-module-script", moduleId);
        script.onload = resolve;
        script.onerror = reject;
        document.body.appendChild(script);
    });
}
