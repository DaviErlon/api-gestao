// Dados dos módulos (conteúdo estático por enquanto)
const modules = {
    posts: {
        title: "Posts",
        description: "Módulo de Posts: gerencie artigos, notícias e publicações. (Em breve funcionalidades completas)"
    },
    reunioes: {
        title: "Reuniões",
        description: "Módulo de Reuniões: agende, visualize e participe de reuniões. (Em construção)"
    },
    ciclos: {
        title: "Ciclos",
        description: "Módulo de Ciclos: acompanhe ciclos de trabalho, metas e feedbacks. (Em desenvolvimento)"
    }
};

// Carrega o módulo selecionado
function loadModule(moduleId) {
    const module = modules[moduleId];
    if (!module) return;

    const titleEl = document.getElementById("moduleTitle");
    const descEl = document.getElementById("moduleDesc");
    if (titleEl && descEl) {
        titleEl.textContent = module.title;
        descEl.textContent = module.description;
    }

    // Atualiza classe 'active' nos itens da sidebar
    document.querySelectorAll(".nav-item").forEach(item => {
        if (item.getAttribute("data-module") === moduleId) {
            item.classList.add("active");
        } else {
            item.classList.remove("active");
        }
    });
}

// Configura eventos da sidebar
function setupSidebar() {
    const navItems = document.querySelectorAll(".nav-item");
    navItems.forEach(item => {
        item.addEventListener("click", () => {
            const moduleId = item.getAttribute("data-module");
            if (moduleId) loadModule(moduleId);
        });
    });
}

// Logout
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

// Exibir avatar com iniciais do usuário (se disponível no JWT)
function setUserAvatar() {
    const payloadStr = localStorage.getItem("user_payload");
    if (payloadStr) {
        try {
            const payload = JSON.parse(payloadStr);
            let nome = payload.name || payload.email || payload.username || "U";
            if (nome.includes("@")) nome = nome[0].toUpperCase();
            else nome = nome.charAt(0).toUpperCase();
            const avatarDiv = document.getElementById("userAvatar");
            if (avatarDiv) avatarDiv.textContent = nome;
        } catch (e) {}
    }
}

// Inicialização
document.addEventListener("DOMContentLoaded", () => {
    setupSidebar();
    setupLogout();
    setUserAvatar();
    loadModule("posts"); // módulo inicial
});