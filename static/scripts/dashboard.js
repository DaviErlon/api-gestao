const modules = {
    posts: {
        page: "/posts.html",
        script: "/scripts/posts.js",
        style: "/styles/posts.css",
        init: "initPosts"
    },
    reunioes: {
        page: "/reunioes.html",
        script: "/scripts/reunioes.js",
        style: "/styles/reunioes.css",
        init: "initReunioes"
    },
    empresas: {
        page: "/empresas.html",
        script: "/scripts/empresas.js",
        style: "/styles/empresas.css",
        init: "initEmpresas"
    },
    ciclos: {
        page: "/ciclos.html",
        script: "/scripts/ciclos.js",
        style: "/styles/ciclos.css",
        init: "initCiclos"
    },
    profile: {
        page: "/profile.html",
        script: "/scripts/profile.js",
        style: "/styles/profile.css",
        init: "initProfile"
    }
};

// Carrega script e aguarda estar pronto
function loadScript(src, moduleId) {
    return new Promise((resolve, reject) => {
        // Remove scripts anteriores do módulo
        document.querySelectorAll("script[data-module-script]").forEach(el => el.remove());

        const script = document.createElement("script");
        // ✅ cache-busting: força o browser a baixar a versão mais recente
        script.src = src + "?v=" + Date.now();
        script.setAttribute("data-module-script", moduleId);
        script.onload = resolve;
        script.onerror = reject;
        document.body.appendChild(script);
    });
}

async function loadModule(moduleId) {
    const module = modules[moduleId];
    if (!module) return;

    const contentEl = document.getElementById("mainContent");

    try {
        // 1. Carrega HTML
        const response = await fetch(module.page);
        if (!response.ok) throw new Error("Falha ao carregar HTML do módulo");
        const html = await response.text();
        contentEl.innerHTML = html;

        // 2. Troca CSS — com cache-busting
        document.querySelectorAll("link[data-module-style]").forEach(el => el.remove());
        if (module.style) {
            const link = document.createElement("link");
            link.rel = "stylesheet";
            // ✅ cache-busting: garante que o CSS novo seja carregado
            link.href = module.style + "?v=" + Date.now();
            link.setAttribute("data-module-style", moduleId);
            document.head.appendChild(link);
        }

        // 3. Carrega script e AGUARDA antes de chamar o init
        if (module.script) {
            await loadScript(module.script, moduleId);
        }

        // 4. Chama a função de init do módulo
        if (module.init && typeof window[module.init] === "function") {
            window[module.init]();
        }

    } catch (err) {
        console.error(`Erro ao carregar módulo "${moduleId}":`, err);
        contentEl.innerHTML = "<p>Erro ao carregar módulo</p>";
    }

    // Atualiza menu ativo
    document.querySelectorAll(".nav-item").forEach(item => {
        item.classList.toggle("active", item.getAttribute("data-module") === moduleId);
    });
}

function setupProfileClick() {
    const avatar = document.getElementById("userAvatar");
    if (!avatar) return;

    avatar.addEventListener("click", () => {
        loadModule("profile");
    });
}

function setupSidebar() {
    document.querySelectorAll(".nav-item").forEach(item => {
        item.addEventListener("click", () => {
            const moduleId = item.getAttribute("data-module");
            if (moduleId) loadModule(moduleId);
        });
    });
}

function setupLogout() {
    const logoutBtn = document.getElementById("logoutBtn");
    if (logoutBtn) {
        logoutBtn.addEventListener("click", () => {
            localStorage.removeItem("auth_token");
            localStorage.removeItem("user_payload");
            window.location.href = "/login.html";
        });
    }
}

function setUserAvatar() {
    const payloadStr = localStorage.getItem("user_payload");
    if (!payloadStr) return;

    try {
        const payload = JSON.parse(payloadStr);
        let nome = payload.name || payload.email || payload.username || "U";
        nome = nome.includes("@") ? nome[0].toUpperCase() : nome.charAt(0).toUpperCase();

        const avatarDiv = document.getElementById("userAvatar");
        if (avatarDiv) avatarDiv.textContent = nome;
    } catch (e) { }
}

document.addEventListener("DOMContentLoaded", () => {
    setupSidebar();
    setupLogout();
    setupProfileClick()
    setUserAvatar();
    loadModule("posts");
});