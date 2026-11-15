const state = {
    id: Number(new URLSearchParams(window.location.search).get("id")),
    title: null,
    status: { saved: false, watched: false, stars: 0, review: "" },
    myStars: null,
    reviewStars: null
};

function showMessage(message) {
    document.querySelector("#titlePage").hidden = true;
    const element = document.querySelector("#pageMessage");
    element.textContent = message;
    element.hidden = false;
}

function scoreHTML(title) {
    if (title.rating_average == null) {
        return `<div><strong>No ratings yet</strong></div>`;
    }

    return `
        <span class="score-num">${title.rating_average.toFixed(1)}</span>
        <div>
            <div class="stars-static">${starText(Math.round(title.rating_average))}</div>
        </div>
    `;
}

function renderHero() {
    const { title } = state;
    const hero = document.querySelector("#titleHero");
    const genres = (title.genres ?? [])
        .map((genre) => `<a class="chip" href="browse.html?genre=${encodeURIComponent(genre.name)}">${escapeHTML(genre.name)}</a>`)
        .join("");

    hero.innerHTML = `
        <img class="hero-bg" src="${escapeHTML(posterURL(title))}" alt="">
        <div class="hero-content">
            ${posterImage(title, "hero-poster")}
            <div>
                <h1>${escapeHTML(title.name)}</h1>
                ${genres ? `<div class="chip-row">${genres}</div>` : ""}
                <p class="hero-text">${escapeHTML(title.synopsis || "No description.")}</p>

                <div class="score-row">
                    <div id="communityScore" class="score">${scoreHTML(title)}</div>
                    <div class="your-rating">
                        <span>Your rating</span>
                        <div id="myStars" class="stars"></div>
                    </div>
                </div>

                <div class="actions">
                    <button class="btn btn-primary" type="button" data-save="${title.id}"></button>
                    <button id="watchedButton" class="btn" type="button"></button>
                </div>
            </div>
        </div>
    `;

    hero.querySelectorAll("[data-save]").forEach(updateSaveButton);
    updateWatchedButton();

    const myStars = hero.querySelector("#myStars");
    if (session.user) {
        state.myStars = starPicker(myStars, state.status.stars, async (stars) => {
            await saveRating(stars, state.status.review);
        });
    } else {
        myStars.innerHTML = `<a class="btn btn-small" href="signin.html">Sign in to rate</a>`;
    }

    hero.querySelector("#watchedButton").addEventListener("click", toggleWatched);
}

function updateWatchedButton() {
    const button = document.querySelector("#watchedButton");
    button.setAttribute("aria-pressed", state.status.watched);
    button.textContent = state.status.watched ? "✓ Watched" : "Mark as watched";
}

async function toggleWatched(event) {
    if (!session.user) {
        goToSignIn();
        return;
    }

    const button = event.currentTarget;
    button.disabled = true;
    try {
        await api(`/me/watched/${state.id}`, { method: state.status.watched ? "DELETE" : "PUT" });
        state.status.watched = !state.status.watched;
        updateWatchedButton();
    } catch (error) {
        alert(error.message);
    } finally {
        button.disabled = false;
    }
}

async function refreshScore() {
    state.title = await api(`/titles/${state.id}`);
    document.querySelector("#communityScore").innerHTML = scoreHTML(state.title);
}

async function saveRating(stars, review) {
    await api(`/me/ratings/${state.id}`, { method: "PUT", body: { stars, review } });
    state.status.stars = stars;
    state.status.review = review;
    state.status.watched = true;
    updateWatchedButton();
    state.myStars?.set(stars);
    state.reviewStars?.set(stars);
    document.querySelector("#deleteRating").hidden = false;
    await Promise.all([refreshScore(), loadReviews()]);
}

function renderWhereToWatch() {
    const services = (state.title.services ?? [])
        .filter((service) => !/channel/i.test(service.name))
        .sort((a, b) => (isMyService(b) - isMyService(a)) || a.name.localeCompare(b.name));

    document.querySelector("#whereToWatch").innerHTML = services.length
        ? services.map((service) => `
            <li class="${isMyService(service) ? "mine" : ""}">
                <div>
                    <strong>${escapeHTML(service.name)}</strong>
                    <span>${isMyService(service) ? "Included with your plan" : "Subscription"}</span>
                </div>
            </li>
        `).join("")
        : `<li><div><strong>Not on a subscription service</strong><span>No US streaming subscription carries this right now.</span></div></li>`;
}

function renderTrailer() {
    const hasTrailer = Boolean(state.title.trailer_url);
    document.querySelector("#trailer").hidden = !hasTrailer;
    document.querySelector("#titleGrid").classList.toggle("single", !hasTrailer);
    if (hasTrailer) {
        document.querySelector("#trailerFrame").src = state.title.trailer_url;
    }
}

async function loadReviews() {
    const reviews = await api(`/titles/${state.id}/reviews`).catch(() => []);
    document.querySelector("#reviewList").innerHTML = reviews.length
        ? reviews.map((review) => `
            <article class="review">
                <div class="review-head">
                    <span class="avatar small">${escapeHTML(review.username[0].toUpperCase())}</span>
                    <strong>@${escapeHTML(review.username)}</strong>
                    <span class="stars-static">${starText(review.stars)}</span>
                </div>
                <p>${escapeHTML(review.review)}</p>
            </article>
        `).join("")
        : `<p class="extraText">No written reviews yet.</p>`;
}

function setupReviewForm() {
    const form = document.querySelector("#reviewForm");
    const status = document.querySelector("#reviewStatus");
    const deleteButton = document.querySelector("#deleteRating");

    if (!session.user) {
        document.querySelector("#reviewSignIn").hidden = false;
        return;
    }

    form.hidden = false;
    form.elements.review.value = state.status.review;
    deleteButton.hidden = state.status.stars === 0;
    state.reviewStars = starPicker(document.querySelector("#reviewStars"), state.status.stars, () => {});

    form.addEventListener("submit", async (event) => {
        event.preventDefault();
        if (!state.reviewStars.value) {
            setStatus(status, "Pick a rating first.", true);
            return;
        }

        const button = form.querySelector('button[type="submit"]');
        button.disabled = true;
        try {
            await saveRating(state.reviewStars.value, form.elements.review.value.trim());
            setStatus(status, "Review posted.");
        } catch (error) {
            setStatus(status, error.message, true);
        } finally {
            button.disabled = false;
        }
    });

    deleteButton.addEventListener("click", async () => {
        if (!confirm("Delete your review?")) {
            return;
        }

        try {
            await api(`/me/ratings/${state.id}`, { method: "DELETE" });
            state.status.stars = 0;
            state.status.review = "";
            form.elements.review.value = "";
            state.myStars?.set(0);
            state.reviewStars.set(0);
            deleteButton.hidden = true;
            setStatus(status, "Rating deleted.");
            await Promise.all([refreshScore(), loadReviews()]);
        } catch (error) {
            setStatus(status, error.message, true);
        }
    });
}

async function setupMovie() {
    if (!Number.isInteger(state.id) || state.id <= 0) {
        showMessage("Title not found.");
        return;
    }

    try {
        state.title = await api(`/titles/${state.id}`);
    } catch (error) {
        showMessage(error.status === 404 ? "Title not found." : error.message);
        return;
    }

    await sessionReady;

    const [status, similar] = await Promise.all([
        session.user ? api(`/me/titles/${state.id}`).catch(() => null) : null,
        api(`/titles/${state.id}/similar`).catch(() => [])
    ]);

    if (status) {
        state.status = status;
    }

    document.title = `7Movies | ${state.title.name}`;
    document.querySelector("#pageMessage").hidden = true;
    document.querySelector("#titlePage").hidden = false;

    renderHero();
    renderWhereToWatch();
    renderTrailer();
    renderCards(document.querySelector("#similar"), similar);
    document.querySelector("#similarSection").hidden = similar.length === 0;
    setupReviewForm();
    await loadReviews();

    if (window.location.hash) {
        document.getElementById(window.location.hash.slice(1))?.scrollIntoView();
    }
}

setupMovie();
