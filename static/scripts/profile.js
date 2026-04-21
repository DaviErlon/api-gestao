const PROFILE_API = "/api/user/users";

// =========================================================
// INIT
// =========================================================

function initProfile() {
    loadProfile();
    setupProfileForm();
}

window.initProfile = initProfile;

function showMessage(text, type = "error") {
    const el = document.getElementById("profileMessage");
    if (!el) return;

    el.textContent = text;
    el.className = `message ${type}`;

    // fade-in simples
    el.style.opacity = "1";

    // opcional: sumir depois de 4s
    setTimeout(() => {
        el.style.opacity = "0.8";
    }, 4000);
}

// =========================================================
// GET USER
// =========================================================

async function loadProfile() {
    const payload = JSON.parse(localStorage.getItem("user_payload") || "{}");
    const userId = payload.user_id;

    if (!userId) return;

    try {
        const res = await fetch(`${PROFILE_API}/${userId}`, {
            headers: {
                "Authorization": "Bearer " + localStorage.getItem("auth_token")
            }
        });

        if (!res.ok) throw new Error("Erro ao carregar perfil");

        const user = await res.json();

        document.getElementById("profileNameInput").value = user.name || "";
        document.getElementById("profileLoginInput").value = user.login || "";

        const avatar = document.getElementById("profileAvatar");
        avatar.textContent = user.name?.[0]?.toUpperCase() || "U";

    } catch (err) {
        console.error(err);
        showMessage("Erro ao carregar perfil");
    }
}

// =========================================================
// FORM
// =========================================================

function setupProfileForm() {
    const form = document.getElementById("profileForm");
    const btn = document.getElementById("saveProfileBtn");
    const currPasswordInput = document.getElementById("profileCurrentPassword");

    if (!form) return;

    // 🔥 habilita botão só se tiver senha atual
    currPasswordInput.addEventListener("input", () => {
        btn.disabled = !currPasswordInput.value.trim();
    });

    form.onsubmit = async (e) => {
        e.preventDefault();

        const name = document.getElementById("profileNameInput").value.trim();
        const login = document.getElementById("profileLoginInput").value.trim();
        const password = document.getElementById("profilePasswordInput").value;
        const currPassword = currPasswordInput.value;

        btn.disabled = true;

        try {
            const res = await fetch(PROFILE_API, {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": "Bearer " + localStorage.getItem("auth_token")
                },
                body: JSON.stringify({
                    name,
                    login,
                    password: password || undefined,
                    curr_password: currPassword
                })
            });

            if (!res.ok) {
                // tenta pegar mensagem do backend
                let errMsg = "Erro ao atualizar perfil";

                try {
                    const data = await res.json();
                    errMsg = data.error || errMsg;
                } catch { }

                throw new Error(errMsg);
            }

            showMessage("Perfil atualizado com sucesso", "success");

            // limpa campos sensíveis
            document.getElementById("profilePasswordInput").value = "";
            currPasswordInput.value = "";

        } catch (err) {
            console.error(err);
            showMessage(err.message || "Erro inesperado");
        } finally {
            btn.disabled = true;
        }
    };
}