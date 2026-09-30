// requestAPI sends JSON requests using the browser's session cookie.
async function requestAPI(path, method = "GET", body) {
    const response = await fetch(`/api${path}`, {
        method,
        credentials: "same-origin",
        cache: "no-store",
        headers: body ? { "Content-Type": "application/json" } : {},
        body: body ? JSON.stringify(body) : undefined
    });

    if (response.status === 204) {
        return;
    }

    if (!response.headers.get("Content-Type")?.includes("application/json")) {
        throw new Error("The account service is unavailable. Please try again later.");
    }

    const data = await response.json();
    if (!response.ok) {
        const error = new Error(data.error || "The request failed. Please try again.");
        error.status = response.status;
        throw error;
    }

    return data;
}

// setupAuthForm submits registration or login and displays errors in the form.
function setupAuthForm(form, path, destination) {
    if (!form) {
        return;
    }

    form.addEventListener("submit", async (event) => {
        event.preventDefault();
        const button = form.querySelector('button[type="submit"]');
        const message = form.querySelector(".form-message");
        button.disabled = true;
        message.textContent = "";

        try {
            await requestAPI(path, "POST", Object.fromEntries(new FormData(form)));
            window.location.assign(destination);
        } catch (error) {
            message.textContent = error instanceof TypeError ? "Unable to reach the account service. Please try again." : error.message;
        } finally {
            button.disabled = false;
        }
    });
}

// setupSession displays the current account and connects the sign-out links.
async function setupSession() {
    const links = document.querySelectorAll("[data-session-link]");
    const profileName = document.querySelector("#profileName");
    if (!links.length && !profileName) {
        return;
    }

    try {
        const user = await requestAPI("/me");
        if (profileName) {
            profileName.textContent = `@${user.username}`;
            document.querySelector("#profileEmail").textContent = user.email;
        }

        links.forEach((link) => {
            link.textContent = "Sign Out";
            link.addEventListener("click", async (event) => {
                event.preventDefault();
                if (link.getAttribute("aria-disabled") === "true") {
                    return;
                }

                link.setAttribute("aria-disabled", "true");

                try {
                    await requestAPI("/sessions/current", "DELETE");
                    window.location.assign("signin.html");
                } catch (error) {
                    document.querySelector("#sessionMessage").textContent = "Could not sign out. Please try again.";
                } finally {
                    link.removeAttribute("aria-disabled");
                }
            });
        });
    } catch (error) {
        if (error.status === 401) {
            if (profileName) {
                window.location.replace("signin.html");
            }
        } else {
            document.querySelector("#sessionMessage").textContent = "Unable to load your account. Please try again later.";
            if (profileName) {
                profileName.textContent = "Profile unavailable";
            }
        }
    }
}

setupAuthForm(document.querySelector("#signinForm"), "/sessions", "profile.html");
setupAuthForm(document.querySelector("#createForm"), "/users", "signin.html?registered=1");

if (document.querySelector("#signinForm") && new URLSearchParams(window.location.search).get("registered") === "1") {
    document.querySelector("#authMessage").textContent = "Account created. Sign in to continue.";
}

setupSession();
