(function() {
const API_EMPRESAS = "/api/user/empresas";

// =========================================================
// INIT
// =========================================================

function initEmpresas() {
    loadEmpresas();
}

window.initEmpresas = initEmpresas;

// =========================================================
// FETCH
// =========================================================

async function loadEmpresas() {
    const container = document.getElementById("empresasContainer");

    try {
        const res = await fetch(API_EMPRESAS, {
            headers: authHeaders()
        });

        if (!res.ok) {
            let msg = "Erro ao carregar empresas";

            try {
                const data = await res.json();
                msg = data.error || msg;
            } catch {
                msg = await res.text();
            }

            throw new Error(msg);
        }

        const empresas = await res.json();

        container.innerHTML = "";

        if (!empresas.length) {
            showMessage("empresasMessage", "Nenhuma empresa encontrada", "info");
            return;
        }

        empresas.forEach(emp => {
            container.appendChild(createEmpresaCard(emp));
        });

    } catch (err) {
        console.error(err);
        showMessage("empresasMessage", err.message, "error");
    }
}

// =========================================================
// CARD
// =========================================================

function createEmpresaCard(emp) {
    const div = document.createElement("div");
    div.className = "empresa-card";

    div.innerHTML = `
        <div class="empresa-name" title="${emp.name}">
            ${escapeHtml(emp.name)}
        </div>
    `;

    return div;
}

})();
