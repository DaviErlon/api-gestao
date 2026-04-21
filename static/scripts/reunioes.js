/* ============================================================
   reunioes.js — módulo de reuniões
   ============================================================ */

const defaultReunioes = [
    { id: 1, title: 'Planejamento semanal', date: '20/04/2026 10:00', status: 'Agendada' },
    { id: 2, title: 'Reunião de resultados', date: '22/04/2026 14:00', status: 'Confirmada' }
];

function initReunioes() {
    const container = document.getElementById('reunioesContainer');
    const actionButton = document.getElementById('createReuniaoBtn');

    if (!container) return;
    renderReunioes(container, defaultReunioes);

    if (actionButton) {
        actionButton.addEventListener('click', () => {
            const title = prompt('Título da reunião:');
            if (!title) return;

            const date = prompt('Data e hora:', '25/04/2026 09:00');
            if (!date) return;

            defaultReunioes.push({
                id: Date.now(),
                title,
                date,
                status: 'Pendente'
            });

            renderReunioes(container, defaultReunioes);
        });
    }
}

window.initReunioes = initReunioes;

function renderReunioes(container, list) {
    container.innerHTML = '';

    if (!list.length) {
        container.innerHTML = '<p class="module-description">Ainda não há reuniões agendadas.</p>';
        return;
    }

    list.forEach(item => {
        const card = document.createElement('div');
        card.className = 'item-card';
        card.innerHTML = `
            <div class="item-card-title">${escapeHtml(item.title)}</div>
            <div class="item-card-meta">${escapeHtml(item.date)} · ${escapeHtml(item.status)}</div>
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
