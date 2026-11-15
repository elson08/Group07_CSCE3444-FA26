const popularServices = [
    "Netflix",
    "Hulu",
    "Amazon Prime Video",
    "HBO Max",
    "Disney Plus",
    "Paramount Plus Essential",
    "Paramount Plus Premium",
    "Peacock Premium",
    "Peacock Premium Plus",
    "Apple TV"
];

const profile = {
    watched: [],
    ratings: [],
    services: []
};

function renderHead() {
    const user = session.user;
    document.querySelector("#profileInitial").textContent = user.username[0].toUpperCase();
    document.querySelector("#profileName").textContent = `@${user.username}`;
    const since = new Date(user.created_at).toLocaleDateString(undefined, { month: "long", year: "numeric" });
    document.querySelector("#profileMeta").textContent = `${user.email} · Member since ${since}`;
}

function renderStats() {
    const average = profile.ratings.length
        ? (profile.ratings.reduce((sum, rating) => sum + rating.stars, 0) / profile.ratings.length).toFixed(1)
        : "–";

    document.querySelector("#profileStats").innerHTML = [
        [session.saved.size, "Watchlist"],
        [profile.watched.length, "Watched"],
        [profile.ratings.length, "Ratings"],
        [average, "Avg rating"]
    ].map(([value, label]) => `<div><strong>${value}</strong><span>${label}</span></div>`).join("");
}

function renderWatchlist() {
    const sort = document.querySelector("#watchlistSort").value;
    const titles = [...session.savedTitles];

    if (sort === "az") {
        titles.sort((a, b) => a.name.localeCompare(b.name));
    } else if (sort === "services") {
        const onMine = (title) => (title.services ?? []).some(isMyService);
        titles.sort((a, b) => onMine(b) - onMine(a));
    }

    renderCards(document.querySelector("#watchlistGrid"), titles, "Your watchlist is empty. Search above or tap + on any title to add it.");
}

function renderWatched() {
    renderCards(document.querySelector("#watchedGrid"), profile.watched, "Nothing here yet. Mark titles as watched or rate them to fill this list.");
}

function renderRatings() {
    const list = document.querySelector("#ratingList");
    list.innerHTML = profile.ratings.length
        ? profile.ratings.map(({ title, stars, review, updated_at: updated }) => `
            <article class="rating-row">
                <img src="${escapeHTML(posterURL(title))}" alt="" loading="lazy" onerror="this.onerror=null;this.src='${placeholderPoster}'">
                <div>
                    <h3><a href="movie.html?id=${title.id}">${escapeHTML(title.name)}</a> <span class="extraText">${title.release_year ? `(${title.release_year})` : ""}</span></h3>
                    <span class="stars-static">${starText(stars)}</span>
                    <p>${escapeHTML(review || "No review written")} · ${escapeHTML(formatDate(updated))}</p>
                </div>
                <div class="rating-actions">
                    <a class="btn btn-small" href="movie.html?id=${title.id}#reviews">Edit</a>
                    <button class="btn btn-small btn-ghost" type="button" data-delete-rating="${title.id}">Delete</button>
                </div>
            </article>
        `).join("")
        : `<p class="no-results">You haven't rated anything yet. Open any title and pick a star rating.</p>`;
}

function renderServices() {
    const search = document.querySelector("#serviceSearch").value.trim().toLowerCase();
    const rank = (service) => {
        const popular = popularServices.indexOf(service.name);
        return isMyService(service) ? -1 : popular === -1 ? popularServices.length : popular;
    };

    document.querySelector("#serviceTiles").innerHTML = profile.services
        .filter((service) => service.name.toLowerCase().includes(search))
        .sort((a, b) => rank(a) - rank(b) || a.name.localeCompare(b.name))
        .map((service) => `
            <button class="service-tile" type="button" data-service="${service.id}" aria-pressed="${isMyService(service)}">
                <span class="service-logo">${escapeHTML(service.name.split(/\s+/).map((word) => word[0]).join("").slice(0, 3).toUpperCase())}</span>
                ${escapeHTML(service.name)}
                <span class="tick">✓</span>
            </button>
        `).join("") || `<p class="no-results">No services match.</p>`;
}

function showTab(name) {
    const tabs = [...document.querySelectorAll("[data-tab]")];
    const chosen = tabs.some((tab) => tab.dataset.tab === name) ? name : "watchlist";
    tabs.forEach((tab) => tab.setAttribute("aria-selected", tab.dataset.tab === chosen));
    document.querySelectorAll(".tab-panel").forEach((panel) => {
        panel.hidden = panel.id !== chosen;
    });
}

function setupTabs() {
    document.querySelectorAll("[data-tab]").forEach((tab) => {
        tab.addEventListener("click", () => {
            history.replaceState(null, "", `#${tab.dataset.tab}`);
            showTab(tab.dataset.tab);
        });
    });

    window.addEventListener("hashchange", () => showTab(window.location.hash.slice(1)));
    showTab(window.location.hash.slice(1));
}

async function reloadSaved() {
    session.savedTitles = await api("/me/titles");
    session.saved = new Set(session.savedTitles.map((title) => title.id));
    renderWatchlist();
    renderStats();
}

function setupWatchlistSearch() {
    const input = document.querySelector("#watchlistSearch");
    const results = document.querySelector("#watchlistResults");
    const status = document.querySelector("#watchlistStatus");
    let matches = [];
    let timer = 0;
    let requestNumber = 0;

    const close = () => {
        results.hidden = true;
    };

    input.addEventListener("input", () => {
        clearTimeout(timer);
        const query = input.value.trim();
        setStatus(status, "");

        if (!query) {
            close();
            return;
        }

        timer = setTimeout(async () => {
            const current = ++requestNumber;
            try {
                matches = await api(`/titles?${new URLSearchParams({ q: query, limit: "8", sort: "az" })}`);
            } catch (error) {
                setStatus(status, error.message, true);
                return;
            }

            if (current !== requestNumber) {
                return;
            }

            results.innerHTML = matches.length
                ? matches.map((title, index) => `
                    <button class="profile-search-option" type="button" data-pick="${index}">
                        ${escapeHTML(title.name)} ${title.release_year ? `(${title.release_year})` : ""} · ${kindLabel(title.kind)}${session.saved.has(title.id) ? " · on your watchlist" : ""}
                    </button>
                `).join("")
                : `<p class="extraText">No matching titles found.</p>`;
            results.hidden = false;
        }, 250);
    });

    results.addEventListener("click", async (event) => {
        const option = event.target.closest("[data-pick]");
        if (!option) {
            return;
        }

        const title = matches[Number(option.dataset.pick)];
        close();
        input.value = "";

        if (session.saved.has(title.id)) {
            setStatus(status, `${title.name} is already on your watchlist.`);
            return;
        }

        try {
            await api(`/me/titles/${title.id}`, { method: "PUT" });
            await reloadSaved();
            setStatus(status, `${title.name} was added to your watchlist.`);
        } catch (error) {
            setStatus(status, error.message, true);
        }
    });

    input.addEventListener("keydown", (event) => {
        if (event.key === "Escape") {
            close();
        }
    });

    document.addEventListener("click", (event) => {
        if (!event.target.closest(".search-add")) {
            close();
        }
    });
}

function setupLists() {
    document.querySelector("#watchlistSort").addEventListener("change", renderWatchlist);
    document.addEventListener("watchlist-change", () => reloadSaved().catch(() => {}));

    document.querySelector("#ratingList").addEventListener("click", async (event) => {
        const button = event.target.closest("[data-delete-rating]");
        if (!button || !confirm("Delete this rating and review?")) {
            return;
        }

        button.disabled = true;
        try {
            await api(`/me/ratings/${button.dataset.deleteRating}`, { method: "DELETE" });
            profile.ratings = await api("/me/ratings");
            renderRatings();
            renderStats();
        } catch (error) {
            alert(error.message);
            button.disabled = false;
        }
    });

    document.querySelector("#serviceSearch").addEventListener("input", renderServices);

    document.querySelector("#serviceTiles").addEventListener("click", async (event) => {
        const tile = event.target.closest("[data-service]");
        if (!tile) {
            return;
        }

        const id = Number(tile.dataset.service);
        const selected = session.services.some((service) => service.id === id);
        tile.disabled = true;
        try {
            await api(`/me/services/${id}`, { method: selected ? "DELETE" : "PUT" });
            session.services = (await api("/me/services")) ?? [];
            renderServices();
            renderWatchlist();
            renderWatched();
        } catch (error) {
            alert(error.message);
            tile.disabled = false;
        }
    });
}

function setupAccount() {
    const details = document.querySelector("#detailsForm");
    details.elements.username.value = session.user.username;
    details.elements.email.value = session.user.email;

    details.addEventListener("submit", async (event) => {
        event.preventDefault();
        const status = details.querySelector(".form-status");
        try {
            session.user = await api("/me", {
                method: "PATCH",
                body: { username: details.elements.username.value.trim(), email: details.elements.email.value.trim() }
            });
            renderHead();
            renderAccountSlot();
            setStatus(status, "Your account details were saved.");
        } catch (error) {
            setStatus(status, error.message, true);
        }
    });

    const password = document.querySelector("#passwordForm");
    password.addEventListener("submit", async (event) => {
        event.preventDefault();
        const status = password.querySelector(".form-status");
        if (password.elements.next.value !== password.elements.confirm.value) {
            setStatus(status, "The new passwords do not match.", true);
            return;
        }

        try {
            await api("/me/password", {
                method: "PUT",
                body: { current_password: password.elements.current.value, new_password: password.elements.next.value }
            });
            password.reset();
            setStatus(status, "Your password was changed.");
        } catch (error) {
            setStatus(status, error.message, true);
        }
    });

    const remove = document.querySelector("#deleteForm");
    remove.addEventListener("submit", async (event) => {
        event.preventDefault();
        const status = remove.querySelector(".form-status");
        if (!confirm("Permanently delete your account? This cannot be undone.")) {
            return;
        }

        try {
            await api("/me", { method: "DELETE", body: { password: remove.elements.password.value } });
            window.location.assign("index.html");
        } catch (error) {
            setStatus(status, error.message, true);
        }
    });

    document.querySelector("#signOut").addEventListener("click", async () => {
        try {
            await api("/sessions/current", { method: "DELETE" });
        } finally {
            window.location.assign("index.html");
        }
    });
}

async function setupProfile() {
    await sessionReady;
    if (!session.user) {
        window.location.replace("signin.html");
        return;
    }

    try {
        [profile.watched, profile.ratings, profile.services] = await Promise.all([
            api("/me/watched"),
            api("/me/ratings"),
            api("/services")
        ]);
    } catch (error) {
        document.querySelector("#pageMessage").textContent = `Could not load your profile: ${error.message}`;
        return;
    }

    document.querySelector("#pageMessage").hidden = true;
    document.querySelector("#profilePage").hidden = false;

    renderHead();
    renderStats();
    renderWatchlist();
    renderWatched();
    renderRatings();
    renderServices();
    setupTabs();
    setupWatchlistSearch();
    setupLists();
    setupAccount();
}

setupProfile();
