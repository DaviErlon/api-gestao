/* ============================================================
   ciclos.js — módulo de ciclos
   ============================================================ */

const defaultCiclos = [
    { id: 1, name: 'Ciclo Alpha', progress: '45%', phase: 'Análise' },
    { id: 2, name: 'Ciclo Beta', progress: '75%', phase: 'Entrega' }
];

function initCiclos() {
    const container = document.getElementById('ciclosContainer');
    const actionButton = document.getElementById('createCicloBtn');

    if (!container) return;
    renderCiclos(container, defaultCiclos);

    if (actionButton) {
        actionButton.addEventListener('click', () => {
            const name = prompt('Nome do ciclo:');
            if (!name) return;

            defaultCiclos.push({
                id: Date.now(),
                name,
                progress: '0%',
                phase: 'Planejamento'
            });

            renderCiclos(container, defaultCiclos);
        });
    }
}

window.initCiclos = initCiclos;

function renderCiclos(container, list) {
    container.innerHTML = '';

    if (!list.length) {
        container.innerHTML = '<p class="module-description">Nenhum ciclo registrado no momento.</p>';
        return;
    }

    list.forEach(item => {
        const card = document.createElement('div');
        card.className = 'item-card';
        card.innerHTML = `
            <div class="item-card-title">${escapeHtml(item.name)}</div>
            <div class="item-card-meta">${escapeHtml(item.progress)} · ${escapeHtml(item.phase)}</div>
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
