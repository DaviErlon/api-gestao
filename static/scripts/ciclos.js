(function() {
const API_CICLOS = "/api/user/ciclos";
const API_DECISOES = "/api/user/decisoes";

// =========================================================
// STATE
// =========================================================
let ciclos = [];
let decisoesMap = {}; // Map ciclo_id -> decisao para acesso rápido
let cicloAtualParaDecisao = null; // Armazena ciclo atual ao criar decisão

// =========================================================
// INIT
// =========================================================
function initCiclos() {
    setupViewCicloModal();
    setupCreateCicloModal();
    setupCreateDecisaoModal();
    loadCiclosWithDecisoes();
}

window.initCiclos = initCiclos;

// =========================================================
// LISTAR CICLOS (SEM PAGINAÇÃO)
// =========================================================
async function loadCiclosWithDecisoes() {
    const container = document.getElementById("ciclosContainer");

    try {
        // Carregar ciclos E decisões em paralelo
        const [ciclosRes, decisoesRes] = await Promise.all([
            fetch(API_CICLOS, {
                method: "GET",
                headers: authHeaders()
            }),
            fetch(API_DECISOES, {
                method: "GET",
                headers: authHeaders()
            })
        ]);

        if (!ciclosRes.ok) throw new Error("Erro ao carregar ciclos");
        if (!decisoesRes.ok) throw new Error("Erro ao carregar decisões");

        // Processar ciclos
        ciclos = await ciclosRes.json();

        // Processar decisões e criar mapa
        const decisoes = await decisoesRes.json();
        decisoesMap = {};
        if (Array.isArray(decisoes)) {
            decisoes.forEach(d => {
                decisoesMap[d.ciclo_id] = d;
            });
        }

        // Renderizar ciclos
        container.innerHTML = "";

        if (ciclos.length === 0) {
            container.innerHTML = '<p style="text-align: center; color: var(--color-text-muted); padding: 2rem;">Nenhum ciclo encontrado</p>';
        } else {
            ciclos.forEach(ciclo => {
                container.appendChild(createCicloCard(ciclo));
            });
        }

    } catch (err) {
        console.error(err);
        showMessage("ciclosMessage", "Erro ao carregar ciclos", "error");
    }
}

async function loadCiclos() {
    const container = document.getElementById("ciclosContainer");

    try {
        const res = await fetch(API_CICLOS, {
            method: "GET",
            headers: authHeaders()
        });

        if (!res.ok) throw new Error("Erro ao carregar ciclos");

        ciclos = await res.json();
        container.innerHTML = "";

        if (ciclos.length === 0) {
            container.innerHTML = '<p style="text-align: center; color: var(--color-text-muted); padding: 2rem;">Nenhum ciclo encontrado</p>';
        } else {
            ciclos.forEach(ciclo => {
                container.appendChild(createCicloCard(ciclo));
            });
        }

    } catch (err) {
        console.error(err);
        showMessage("ciclosMessage", "Erro ao carregar ciclos", "error");
    }
}

// =========================================================
// CARREGAR DECISÕES
// =========================================================
async function loadDecisoes() {
    try {
        const res = await fetch(API_DECISOES, {
            method: "GET",
            headers: authHeaders()
        });

        if (!res.ok) throw new Error("Erro ao carregar decisões");

        const decisoes = await res.json();

        // Cria um mapa para acesso O(1)
        decisoesMap = {};
        if (Array.isArray(decisoes)) {
            decisoes.forEach(d => {
                decisoesMap[d.ciclo_id] = d;
            });
        }

    } catch (err) {
        console.error(err);
    }
}

// =========================================================
// CRIAR CARD DE CICLO
// =========================================================
function createCicloCard(c) {
    const div = document.createElement("div");
    div.className = "ciclo-card";

    const decisao = decisoesMap[c.id];

    div.innerHTML = `
        <div class="ciclo-header">
            <div class="ciclo-title">Rodada ${c.rodada}</div>
            ${!decisao ? `<div class="ciclo-id">#${c.id}</div>` : ''}
        </div>

        <div class="ciclo-stats">
            <div class="stat"><strong>Saldo:</strong> ${formatCurrency(c.saldo)}</div>
            <div class="stat"><strong>Dívida:</strong> ${formatCurrency(c.divida)}</div>
            <div class="stat"><strong>Clientes:</strong> ${c.clientes}</div>
            <div class="stat"><strong>NPS:</strong> ${c.nps}</div>
        </div>

        <div class="ciclo-highlight">
            <strong>Valuation:</strong> ${formatCurrency(c.valuation)}
        </div>

        ${decisao ? `<div class="ciclo-decision-indicator">📊 Com decisão</div>` : ''}
    `;

    div.onclick = () => openCicloModal(c, decisao);
    return div;
}

// =========================================================
// MODAL VISUALIZAÇÃO
// =========================================================
function setupViewCicloModal() {
    const modal = document.getElementById("cicloModal");
    const closeBtn = document.getElementById("closeCicloModal");

    if (!modal || !closeBtn) return;

    closeBtn.onclick = () => closeModal("cicloModal");

    modal.onclick = (e) => {
        if (e.target === modal) closeModal("cicloModal");
    };

    // Fechar com ESC
    document.addEventListener("keydown", (e) => {
        if (e.key === "Escape" && !modal.classList.contains("hidden")) {
            closeModal("cicloModal");
        }
    });
}

function openCicloModal(ciclo, decisao) {
    const modal = document.getElementById("cicloModal");
    const titleEl = document.getElementById("cicloModalTitle");
    const contentEl = document.getElementById("cicloModalContent");

    if (!modal) return;

    titleEl.textContent = `Ciclo ${ciclo.rodada}`;

    contentEl.innerHTML = `
        <div class="ciclo-modal-layout">
            <!-- COLUNA ESQUERDA: Dados do Ciclo -->
            <div class="ciclo-modal-left">
                <div class="detail-section">
                    <h4>💰 Financeiro</h4>
                    <div class="detail-row">
                        <span>Saldo: ${ formatCurrency(ciclo.saldo)}</span>
                    </div>
                    <div class="detail-row">
                        <span>Dívida: ${ formatCurrency(ciclo.divida)}</span>
                    </div>
                    <div class="detail-row">
                        <span>Juros: ${ Number(ciclo.juros).toFixed(2)}%</span>
                    </div>
                </div>

                <div class="detail-section">
                    <h4>📊 Operacional</h4>
                    <div class="detail-row">
                        <span>Clientes: ${ ciclo.clientes}</span>
                    </div>
                    <div class="detail-row">
                        <span>Market Share: ${ Number(ciclo.market_share).toFixed(2)}%</span>
                    </div>
                    <div class="detail-row">
                        <span>Tecnologia: ${ ciclo.tech}</span>
                    </div>
                </div>

                <div class="detail-section">
                    <h4>⭐ Reputação & Segurança</h4>
                    <div class="detail-row">
                        <span>Reputação: ${ ciclo.reputacao}</span>
                    </div>
                    <div class="detail-row">
                        <span>Segurança: ${ ciclo.seguranca}</span>
                    </div>
                    <div class="detail-row">
                        <span>NPS: ${ ciclo.nps}</span>
                    </div>
                </div>

                <div class="detail-section highlight">
                    <div class="detail-row">
                        <span>Valuation: ${ formatCurrency(ciclo.valuation)}</span>
                    </div>
                </div>
            </div>

            <!-- COLUNA DIREITA: Decisões ou Empty State -->
            <div class="ciclo-modal-right">
                ${decisao ? `
                    <div class="decision-panel">
                        <h4>📋 Decisões</h4>

                        <div class="decision-grid">
                            <div class="decision-card marketing">
                                <div class="decision-icon">📢</div>
                                <div class="decision-label">Marketing</div>
                                <div class="decision-value">${formatCurrency(decisao.marketing)}</div>
                            </div>

                            <div class="decision-card ped">
                                <div class="decision-icon">🔬</div>
                                <div class="decision-label">P&D</div>
                                <div class="decision-value">${formatCurrency(decisao.ped)}</div>
                            </div>

                            <div class="decision-card suporte">
                                <div class="decision-icon">🛠️</div>
                                <div class="decision-label">Suporte</div>
                                <div class="decision-value">${formatCurrency(decisao.suporte)}</div>
                            </div>

                            <div class="decision-card seguranca">
                                <div class="decision-icon">🔒</div>
                                <div class="decision-label">Segurança</div>
                                <div class="decision-value">${formatCurrency(decisao.seguranca)}</div>
                            </div>

                            <div class="decision-card expansao">
                                <div class="decision-icon">📈</div>
                                <div class="decision-label">Expansão</div>
                                <div class="decision-value">${decisao.expansao}</div>
                            </div>
                        </div>

                        <div class="decision-summary">
                            <div class="summary-row">
                                <span>Total Investido:</span>
                                <strong>${formatCurrency(Number(decisao.marketing) + Number(decisao.ped) + Number(decisao.suporte) + Number(decisao.seguranca))}</strong>
                            </div>
                        </div>
                    </div>
                ` : `
                    <div class="empty-decision-state">
                        <div class="empty-icon">📭</div>
                        <h4>Sem decisões</h4>
                        <p>Este ciclo ainda não possui decisões associadas</p>
                        <button id="addDecisaoBtn" class="add-decisao-btn">➕ Adicionar Decisão</button>
                    </div>
                `}
            </div>
        </div>
    `;

    // Se não houver decisão, adicionar listener ao botão que será renderizado
    if (!decisao) {
        setTimeout(() => {
            const addBtn = document.getElementById("addDecisaoBtn");
            if (addBtn) {
                addBtn.onclick = (e) => {
                    e.stopPropagation();
                    openAddDecisaoModal(ciclo);
                };
            }
        }, 10);
    }

    openModal("cicloModal");
}

// =========================================================
// MODAL CRIAÇÃO
// =========================================================
function setupCreateCicloModal() {
    setupModalToggle("createCicloBtn", "createCicloModal", "cancelCreateCiclo");

    const submitBtn = document.getElementById("submitCreateCiclo");
    if (submitBtn) {
        submitBtn.onclick = async () => {
            await submitCreateCiclo();
        };
    }
}

async function submitCreateCiclo() {
    const submitBtn = document.getElementById("submitCreateCiclo");
    if (!submitBtn) return;

    const fields = {
        rodada: document.getElementById("cicloRodada"),
        saldo: document.getElementById("cicloSaldo"),
        divida: document.getElementById("cicloDivida"),
        juros: document.getElementById("cicloJuros"),
        clientes: document.getElementById("cicloClientes"),
        market_share: document.getElementById("cicloMarket"),
        tech: document.getElementById("cicloTech"),
        reputacao: document.getElementById("cicloReputacao"),
        seguranca: document.getElementById("cicloSeguranca"),
        nps: document.getElementById("cicloNps"),
        valuation: document.getElementById("cicloValuation")
    };

    // Validação
    const payload = {};
    for (const [key, el] of Object.entries(fields)) {
        if (!el) continue;
        const value = el.value.trim();
        if (!value) {
            showMessage("ciclosMessage", `Campo ${key} é obrigatório`, "error");
            return;
        }
        payload[key] = Number(value);
    }

    submitBtn.disabled = true;

    try {
        console.log(payload);
        const res = await fetch(API_CICLOS, {
            method: "POST",
            headers: authHeaders(),
            body: JSON.stringify(payload)
        });

        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.message || "Erro ao criar ciclo");
        }

        closeModal("createCicloModal");

        // Limpar formulário
        Object.values(fields).forEach(el => {
            if (el) el.value = "";
        });

        showMessage("ciclosMessage", "Ciclo criado com sucesso!", "success", 3000);
        loadCiclos();

    } catch (err) {
        console.error(err);
        showMessage("ciclosMessage", err.message, "error");
    } finally {
        submitBtn.disabled = false;
    }
}

// =========================================================
// ABRIR MODAL DE ADICIONAR DECISÃO
// =========================================================
function openAddDecisaoModal(ciclo) {
    cicloAtualParaDecisao = ciclo;

    const modal = document.getElementById("createDecisaoModal");

    if (!modal) {
        showMessage("ciclosMessage", "Erro ao abrir modal", "error");
        return;
    }

    // Encontrar o modal-card dentro do modal-overlay
    let modalCard = modal.querySelector(".modal-card");

    if (!modalCard) {
        // Se não existir, criar
        modalCard = document.createElement("div");
        modalCard.className = "modal-card";
        modal.appendChild(modalCard);
    }

    // Renderizar conteúdo dinâmico apenas no modal-card com IDs únicos
    modalCard.innerHTML = `
        <h3>Adicionar Decisão - Ciclo ${ciclo.rodada}</h3>

        <input data-field="marketing" type="number" placeholder="Marketing" step="0.01" min="0" />
        <input data-field="ped" type="number" placeholder="P&D" step="0.01" min="0" />
        <input data-field="suporte" type="number" placeholder="Suporte" step="0.01" min="0" />
        <input data-field="seguranca" type="number" placeholder="Segurança" step="0.01" min="0" />
        <input data-field="expansao" type="number" placeholder="Expansão (%)" min="0" />

        <div class="modal-actions">
            <button class="modal-cancel" data-action="cancel">Cancelar</button>
            <button class="modal-submit" data-action="submit">Adicionar</button>
        </div>
    `;

    // Adicionar listeners aos botões
    setTimeout(() => {
        const cancelBtn = modalCard.querySelector('[data-action="cancel"]');
        const submitBtn = modalCard.querySelector('[data-action="submit"]');

        if (cancelBtn) {
            cancelBtn.onclick = () => {
                closeModal("createDecisaoModal");
                cicloAtualParaDecisao = null;
            };
        }

        if (submitBtn) {
            submitBtn.onclick = async () => {
                await submitCreateDecisao(modalCard);
            };
        }
    }, 0);

    // Fechar o modal do ciclo e abrir o de decisão
    closeModal("cicloModal");
    openModal("createDecisaoModal");
}

// =========================================================
// MODAL CRIAÇÃO DECISÃO
// =========================================================
function setupCreateDecisaoModal() {
    const modal = document.getElementById("createDecisaoModal");

    if (!modal) return;

    modal.onclick = (e) => {
        if (e.target === modal) closeModal("createDecisaoModal");
    };

    // Fechar com ESC
    document.addEventListener("keydown", (e) => {
        if (e.key === "Escape" && !modal.classList.contains("hidden")) {
            closeModal("createDecisaoModal");
        }
    });
}

async function submitCreateDecisao(modalCard) {
    if (!cicloAtualParaDecisao) {
        showMessage("ciclosMessage", "Ciclo não selecionado", "error");
        return;
    }

    const submitBtn = modalCard?.querySelector('[data-action="submit"]');
    if (!submitBtn) return;

    // Buscar inputs pelo data-field em vez de ID
    const fields = {
        marketing: modalCard?.querySelector('[data-field="marketing"]'),
        ped: modalCard?.querySelector('[data-field="ped"]'),
        suporte: modalCard?.querySelector('[data-field="suporte"]'),
        seguranca: modalCard?.querySelector('[data-field="seguranca"]'),
        expansao: modalCard?.querySelector('[data-field="expansao"]')
    };

    // Validação
    const payload = {};
    for (const [key, el] of Object.entries(fields)) {
        if (!el) continue;
        const value = el.value.trim();
        if (!value) {
            showMessage("ciclosMessage", `Campo ${key} é obrigatório`, "error");
            return;
        }
        payload[key] = Number(value);
    }

    payload.ciclo_id = cicloAtualParaDecisao.id;
    submitBtn.disabled = true;

    try {
        const res = await fetch(API_DECISOES, {
            method: "POST",
            headers: authHeaders(),
            body: JSON.stringify(payload)
        });

        if (!res.ok) {
            const errorText = await res.text();
            throw new Error(errorText || "Erro ao criar decisão");
        }

        showMessage("ciclosMessage", "Decisão criada com sucesso!", "success", 3000);

        // Fechar modal de decisão
        closeModal("createDecisaoModal");

        // Recarregar dados (ciclos e decisões em paralelo)
        await loadCiclosWithDecisoes();

        // Reabrir modal do ciclo com dados atualizados
        const cicloAtualizado = ciclos.find(c => c.id === cicloAtualParaDecisao.id);
        const decisaoNova = decisoesMap[cicloAtualParaDecisao.id];

        if (cicloAtualizado) {
            openCicloModal(cicloAtualizado, decisaoNova);
        }

        cicloAtualParaDecisao = null;

    } catch (err) {
        console.error(err);
        showMessage("ciclosMessage", err.message, "error");
    } finally {
        if (submitBtn) {
            submitBtn.disabled = false;
        }
    }
}

// =========================================================
// HELPERS DE FORMATAÇÃO
// =========================================================
function formatCurrency(value) {
    return Number(value).toLocaleString("pt-BR", {
        style: "currency",
        currency: "BRL"
    });
}

})();
