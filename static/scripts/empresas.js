/* ============================================================
   empresas.js — módulo de empresas
   ============================================================ */

const defaultEmpresas = [
    { id: 1, name: 'Aurora Tech', city: 'São Paulo', status: 'Ativa' },
    { id: 2, name: 'Luna Consultoria', city: 'Rio de Janeiro', status: 'Parceira' }
];

function initEmpresas() {
    const container = document.getElementById('empresasContainer');
    const actionButton = document.getElementById('createEmpresaBtn');

    if (!container) return;
    renderEmpresas(container, defaultEmpresas);

    if (actionButton) {
        actionButton.addEventListener('click', () => {
            const name = prompt('Nome da empresa:');
            if (!name) return;

            const city = prompt('Cidade:');
            if (!city) return;

            defaultEmpresas.push({
                id: Date.now(),
                name,
                city,
                status: 'Nova'
            });

            renderEmpresas(container, defaultEmpresas);
        });
    }
}

window.initEmpresas = initEmpresas;

function renderEmpresas(container, list) {
    container.innerHTML = '';

    if (!list.length) {
        container.innerHTML = '<p class="module-description">Nenhuma empresa cadastrada ainda.</p>';
        return;
    }

    list.forEach(item => {
        const card = document.createElement('div');
        card.className = 'item-card';
        card.innerHTML = `
            <div class="item-card-title">${escapeHtml(item.name)}</div>
            <div class="item-card-meta">${escapeHtml(item.city)} · ${escapeHtml(item.status)}</div>
        `;
        container.appendChild(card);
    });
}

function escapeHtml(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
}
