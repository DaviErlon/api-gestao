async function checkAuth() {
    try {
        const response = await fetch("/api/test", {
            headers: {
                "Authorization": "Bearer " + localStorage.getItem("auth_token")
            }
        });

        const path = window.location.pathname;

        if (response.ok) {
            if (path === "/index.html" || path === "/") {
                window.location.href = "/dashboard.html";
            }

        } else {
            localStorage.removeItem("auth_token");
            window.location.href = "/login.html";
        }

    } catch (error) {
        // fallback 
        localStorage.removeItem("auth_token");
        window.location.href = "/login.html";
    }
}

checkAuth();