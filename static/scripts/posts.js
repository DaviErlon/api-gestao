(function() {
const API_POSTS = "/api/user/posts";

// =========================================================
// STATE PAGINAÇÃO (com persistência)
// =========================================================
let currentPage = parseInt(localStorage.getItem("posts_page")) || 1;
let limit = 10;
let hasNext = false;

// =========================================================
// INIT
// =========================================================

function initPosts() {
    loadPosts(currentPage);
    setupCreatePost();
}

window.initPosts = initPosts;

// =========================================================
// LISTAR POSTS (COM PAGINAÇÃO)
// =========================================================

async function loadPosts(page = 1) {
    const container = document.getElementById("postsContainer");
    const pagination = document.getElementById("pagination");

    try {
        const res = await fetch(`${API_POSTS}?page=${page}&limit=${limit}`, {
            method: "GET",
            headers: authHeaders()
        });

        if (!res.ok) throw new Error("erro ao buscar posts");

        const result = await res.json();

        const posts = result.data;
        const pag = result.pagination;

        currentPage = pag.page;
        hasNext = pag.has_next;

        // 🔥 salva no localStorage
        localStorage.setItem("posts_page", currentPage);

        container.innerHTML = "";

        posts.forEach(post => {
            container.appendChild(createPostElement(post));
        });

        renderPaginationUI(pagination);

    } catch (err) {
        console.error(err);
    }
}

// =========================================================
// PAGINAÇÃO UI
// =========================================================

function renderPaginationUI(container) {
    renderPagination("pagination", currentPage, hasNext, loadPosts);
}

// =========================================================
// MODAL CREATE POST
// =========================================================

function setupCreatePost() {
    const btn = document.getElementById("createPostBtn");
    const modal = document.getElementById("postModal");
    const cancel = document.getElementById("cancelPost");
    const submit = document.getElementById("submitPost");
    const textarea = document.getElementById("postContent");

    if (!btn || !modal) return;

    // Setup toggle modal
    setupModalToggle("createPostBtn", "postModal", "cancelPost");

    submit.onclick = async () => {
        const content = textarea.value.trim();
        if (!content) return;

        submit.disabled = true;

        try {
            await fetch(API_POSTS, {
                method: "POST",
                headers: authHeaders(),
                body: JSON.stringify({
                    content,
                    pin: false
                })
            });

            closeModal("postModal");
            textarea.value = "";

            // 🔥 mantém na mesma página
            loadPosts(currentPage);

        } catch (err) {
            console.error(err);
        } finally {
            submit.disabled = false;
        }
    };

    // Limpa o textarea ao abrir o modal
    modal.addEventListener("click", (e) => {
        if (e.target === modal && !modal.classList.contains("hidden")) {
            textarea.value = "";
            textarea.focus();
        }
    });
}

// =========================================================
// RENDER POST
// =========================================================

function createPostElement(post) {
    const user = getUser();
    const canDelete = user.user_id === post.author_id;

    const initial = post.author_name?.[0]?.toUpperCase() || "?";

    const div = document.createElement("div");
    div.className = "post-wrapper";

    div.innerHTML = `
        <div class="post-avatar">${initial}</div>

        <div class="post">
            <div class="post-header">
                <div class="post-user-info">
                    <span class="post-author">${post.author_name}</span>
                </div>

                ${post.pin ? '<span class="pin-badge">📌 Fixado</span>' : ''}
                <button class="pin-btn ${post.pin ? 'pinned' : ''}">📌</button>
            </div>

            <div class="post-content">${escapeHtml(post.content)}</div>

            <div class="post-footer">
                <span>${formatDate(post.posted_at)}</span>

                <div class="post-actions">
                    ${canDelete ? `<button class="delete-btn">🗑</button>` : ""}

                    <button class="like-btn ${post.curtido ? "liked" : ""}">
                        ${post.curtido ? "❤️" : "🤍"} ${post.likes ?? 0}
                    </button>
                </div>
            </div>
        </div>
    `;

    // LIKE
    const likeBtn = div.querySelector(".like-btn");
    likeBtn.onclick = () => toggleLike(post.id, likeBtn);

    // DELETE
    const deleteBtn = div.querySelector(".delete-btn");
    if (deleteBtn) {
        deleteBtn.onclick = () => deletePost(post.id);
    }

    // PIN
    div.querySelector(".pin-btn").onclick = () => togglePin(post.id);

    return div;
}

// =========================================================
// PIN
// =========================================================

async function togglePin(id) {
    try {
        await fetch(`${API_POSTS}/${id}/pin`, {
            method: "PUT",
            headers: authHeaders()
        });

        loadPosts(currentPage);

    } catch (err) {
        console.error(err);
    }
}

// =========================================================
// LIKE
// =========================================================

async function toggleLike(id, btn) {
    try {
        btn.disabled = true;

        await fetch(`${API_POSTS}/${id}/like`, {
            method: "PUT",
            headers: authHeaders()
        });

        let count = parseInt(btn.innerText.replace(/\D/g, "")) || 0;

        if (btn.classList.contains("liked")) {
            btn.classList.remove("liked");
            btn.innerText = `🤍 ${count - 1}`;
        } else {
            btn.classList.add("liked");
            btn.innerText = `❤️ ${count + 1}`;
        }

    } catch (err) {
        console.error("erro ao dar like", err);
    } finally {
        btn.disabled = false;
    }
}

// =========================================================
// DELETE
// =========================================================

async function deletePost(id) {
    if (!confirm("Deseja deletar este post?")) return;

    try {
        await fetch(`${API_POSTS}/${id}`, {
            method: "DELETE",
            headers: authHeaders()
        });

        loadPosts(currentPage);

    } catch (err) {
        console.error(err);
    }
}

})();

