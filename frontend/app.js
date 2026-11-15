const placeholderPoster = "https://placehold.co/500x750/222c3d/ffffff?text=No+Poster";

function escapeHTML(value) {
    return String(value ?? "")
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#39;");
}

async function request(path, { method = "GET", body } = {}) {
    const response = await fetch(`/api${path}`, {
        method,
        credentials: "same-origin",
        cache: "no-store",
        headers: body === undefined ? {} : { "Content-Type": "application/json" },
        body: body === undefined ? undefined : JSON.stringify(body)
    });

    const type = response.headers.get("Content-Type") || "";
    const data = type.includes("application/json") ? await response.json() : null;

    if (!response.ok) {
        const error = new Error(data?.error || `Request failed (${response.status}).`);
        error.status = response.status;
        throw error;
    }

    return { data, total: Number(response.headers.get("X-Total-Count") || 0) };
}

async function api(path, options) {
    return (await request(path, options)).data;
}

const session = {
    user: null,
    services: [],
    savedTitles: [],
    saved: new Set()
};

async function loadSession() {
    try {
        session.user = await api("/me");
    } catch {
        session.user = null;
    }

    if (session.user) {
        const [services, savedTitles] = await Promise.all([api("/me/services"), api("/me/titles")]);
        session.services = services ?? [];
        session.savedTitles = savedTitles ?? [];
        session.saved = new Set(savedTitles.map((title) => title.id));
    }

    renderAccountSlot();
}

const sessionReady = loadSession();

function renderAccountSlot() {
    document.querySelectorAll("[data-account]").forEach((slot) => {
        slot.innerHTML = session.user
            ? `<a class="avatar" href="profile.html" aria-label="Your profile" title="@${escapeHTML(session.user.username)}">${escapeHTML(session.user.username[0].toUpperCase())}</a>`
            : `<a class="nav-button" href="signin.html">Sign In</a>`;
    });
}

function goToSignIn() {
    window.location.assign("signin.html");
}

function isMyService(service) {
    return session.services.some((mine) => mine.id === service.id);
}

function kindLabel(kind) {
    return kind === "show" ? "Show" : "Movie";
}

function posterURL(title) {
    return title.poster_url || placeholderPoster;
}

function posterImage(title, className = "") {
    return `<img${className ? ` class="${className}"` : ""} src="${escapeHTML(posterURL(title))}" alt="${escapeHTML(title.name)} poster" loading="lazy" onerror="this.onerror=null;this.src='${placeholderPoster}'">`;
}

function starText(count) {
    return "★".repeat(count) + "☆".repeat(5 - count);
}

function titleCard(title) {
    const service = (title.services ?? []).find(isMyService);
    const saved = session.saved.has(title.id);
    const rating = title.rating_average != null
        ? `<span class="badge-rating">★ ${title.rating_average.toFixed(1)}</span>`
        : "";

    return `
        <article class="card">
            <a href="movie.html?id=${title.id}">
                <div class="card-poster">
                    ${posterImage(title)}
                    ${rating}
                    ${service ? `<span class="badge-service">${escapeHTML(service.name)}</span>` : ""}
                </div>
                <div class="card-body">
                    <h3 title="${escapeHTML(title.name)}">${escapeHTML(title.name)}</h3>
                    <p>${title.release_year ?? "Unknown year"}, ${kindLabel(title.kind)}</p>
                </div>
            </a>
            <button class="card-save" type="button" data-save="${title.id}" aria-pressed="${saved}" aria-label="${saved ? "Remove from watchlist" : "Add to watchlist"}">${saved ? "✓" : "+"}</button>
        </article>
    `;
}

function renderCards(container, titles, emptyMessage = "") {
    container.innerHTML = titles.length
        ? titles.map(titleCard).join("")
        : emptyMessage ? `<p class="no-results">${escapeHTML(emptyMessage)}</p>` : "";
}

function updateSaveButton(button) {
    const saved = session.saved.has(Number(button.dataset.save));
    button.setAttribute("aria-pressed", saved);

    if (button.classList.contains("card-save")) {
        button.textContent = saved ? "✓" : "+";
        button.setAttribute("aria-label", saved ? "Remove from watchlist" : "Add to watchlist");
    } else {
        button.textContent = saved ? "✓ In watchlist" : "+ Watchlist";
    }
}

async function toggleSaved(id) {
    const saved = session.saved.has(id);
    await api(`/me/titles/${id}`, { method: saved ? "DELETE" : "PUT" });

    if (saved) {
        session.saved.delete(id);
        session.savedTitles = session.savedTitles.filter((title) => title.id !== id);
    } else {
        session.saved.add(id);
    }

    document.querySelectorAll(`[data-save="${id}"]`).forEach(updateSaveButton);
    document.dispatchEvent(new CustomEvent("watchlist-change", { detail: { id, saved: !saved } }));
}

document.addEventListener("click", async (event) => {
    const button = event.target.closest("[data-save]");
    if (!button) {
        return;
    }

    event.preventDefault();
    await sessionReady;

    if (!session.user) {
        goToSignIn();
        return;
    }

    button.disabled = true;
    try {
        await toggleSaved(Number(button.dataset.save));
    } catch (error) {
        alert(error.message);
    } finally {
        button.disabled = false;
    }
});

function starPicker(container, value, onPick) {
    container.innerHTML = [1, 2, 3, 4, 5]
        .map((star) => `<button type="button" data-star="${star}" aria-label="${star} star${star > 1 ? "s" : ""}">★</button>`)
        .join("") + "<output></output>";

    const output = container.querySelector("output");
    let current = value;

    const paint = (stars) => {
        container.querySelectorAll("[data-star]").forEach((button) => {
            button.classList.toggle("on", Number(button.dataset.star) <= stars);
        });
        output.textContent = stars ? `${stars} / 5` : "";
    };

    paint(current);

    container.addEventListener("click", async (event) => {
        const button = event.target.closest("[data-star]");
        if (!button) {
            return;
        }

        const picked = Number(button.dataset.star);
        const previous = current;
        current = picked;
        paint(picked);

        try {
            await onPick(picked);
        } catch (error) {
            current = previous;
            paint(previous);
            alert(error.message);
        }
    });

    return {
        get value() {
            return current;
        },
        set(stars) {
            current = stars;
            paint(stars);
        }
    };
}

function setStatus(element, message, isError = false) {
    element.textContent = message;
    element.classList.toggle("error", isError);
}
