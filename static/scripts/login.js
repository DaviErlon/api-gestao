// Aguarda o DOM carregar antes de adicionar eventos
document.addEventListener('DOMContentLoaded', () => {
    // Elementos do formulário (IDs do HTML existente)
    const emailInput = document.getElementById('email');
    const senhaInput = document.getElementById('senha');
    const btn = document.getElementById('btn');       // ATENÇÃO: adicione id="btn" ao botão no HTML
    const messageDiv = document.getElementById('message'); // ATENÇÃO: adicione uma div para mensagens

    // Função para ativar/desativar loading no botão
    function setLoading(on) {
        if (!btn) return;
        btn.disabled = on;
        btn.classList.toggle('loading', on);
        // Se quiser mudar o texto do botão durante loading
        if (on) {
            btn.textContent = 'Entrando...';
        } else {
            btn.textContent = 'Entrar';
        }
    }

    // Função para exibir mensagens de feedback
    function showMessage(text, type) {
        if (!messageDiv) return;
        messageDiv.textContent = text;
        messageDiv.className = 'message ' + type; // classes: error, success
    }

    // Decodificar JWT (sem bibliotecas)
    function parseJWT(token) {
        try {
            const payload = token.split('.')[1];
            const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
            return JSON.parse(decoded);
        } catch {
            return null;
        }
    }

    // Função principal de login
    async function handleLogin() {
        const login = emailInput.value.trim();    // usa o email como "login"
        const password = senhaInput.value;

        if (!login || !password) {
            showMessage('Preencha e-mail e senha.', 'error');
            return;
        }

        setLoading(true);
        showMessage(''); // limpa mensagem anterior

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ login, password })
            });

            const data = await response.json();

            if (!response.ok) {
                showMessage(data.message || 'Credenciais inválidas.', 'error');
                setLoading(false);
                return;
            }

            const token = data.token;
            if (!token) {
                showMessage('Resposta inválida do servidor.', 'error');
                setLoading(false);
                return;
            }

            // Salva token e payload do JWT
            localStorage.setItem('auth_token', token);
            const payload = parseJWT(token);
            if (payload) {
                localStorage.setItem('user_payload', JSON.stringify(payload));
            }

            showMessage('Login realizado! Redirecionando...', 'success');
            setTimeout(() => {
                window.location.href = 'dashboard.html';
            }, 800);

        } catch (err) {
            console.error(err);
            showMessage('Erro de conexão. Tente novamente.', 'error');
            setLoading(false);
        }
    }

    // Evento do botão
    if (btn) {
        btn.addEventListener('click', handleLogin);
    }

    // Submeter com a tecla Enter em qualquer campo do formulário
    const formFields = [emailInput, senhaInput];
    formFields.forEach(field => {
        if (field) {
            field.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    handleLogin();
                }
            });
        }
    });
    const toggleSenha = document.getElementById('toggleSenha');

    if (toggleSenha && senhaInput) {
        toggleSenha.addEventListener('click', () => {
            const isPassword = senhaInput.type === 'password';

            senhaInput.type = isPassword ? 'text' : 'password';

            toggleSenha.classList.toggle('active', isPassword);

            // acessibilidade
            toggleSenha.setAttribute(
                'aria-label',
                isPassword ? 'Ocultar senha' : 'Mostrar senha'
            );
        });
    }
});