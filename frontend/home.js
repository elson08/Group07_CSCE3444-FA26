function fillRow(name, titles) {
    const section = document.querySelector(`[data-row="${name}"]`);
    section.hidden = titles.length === 0;
    renderCards(section.querySelector(".scroller"), titles);
}

function renderHero(title) {
    const hero = document.querySelector("#hero");
    const genres = (title.genres ?? []).map((genre) => genre.name).join(", ");
    const meta = [title.release_year, kindLabel(title.kind), genres].filter(Boolean).map(escapeHTML).join(", ");
    const rating = title.rating_average != null
        ? `, <span class="rating">★ ${title.rating_average.toFixed(1)}</span> (${title.rating_count} rating${title.rating_count === 1 ? "" : "s"})`
        : "";

    hero.innerHTML = `
        <img class="hero-bg" src="${escapeHTML(posterURL(title))}" alt="">
        <div class="hero-content">
            ${posterImage(title, "hero-poster")}
            <div>
                <h1>${escapeHTML(title.name)}</h1>
                <p class="meta">${meta}${rating}</p>
                <p class="hero-text">${escapeHTML(title.synopsis || "No description.")}</p>
                <div class="actions">
                    <a class="btn btn-primary" href="movie.html?id=${title.id}">View details</a>
                    <button class="btn" type="button" data-save="${title.id}"></button>
                    ${title.trailer_url ? `<a class="btn btn-ghost" href="movie.html?id=${title.id}#trailer">▶ Trailer</a>` : ""}
                </div>
            </div>
        </div>
    `;

    hero.querySelectorAll("[data-save]").forEach(updateSaveButton);
    hero.hidden = false;
}

function renderServiceBar() {
    const bar = document.querySelector("#serviceBar");
    if (!session.user || !session.services.length) {
        bar.hidden = true;
        return;
    }

    bar.innerHTML = `<span class="label">Show titles on:</span>
           <a class="chip active" href="browse.html?mine=1">All my services</a>
           ${session.services.map((service) => `<a class="chip" href="browse.html?service=${service.id}">${escapeHTML(service.name)}</a>`).join("")}
           <a class="chip" href="profile.html#services">Edit services</a>`;
    bar.hidden = false;
}

function renderGenres(genres) {
    document.querySelector("#genreTiles").innerHTML = genres
        .map((genre) => `<a href="browse.html?genre=${encodeURIComponent(genre.name)}">${escapeHTML(genre.name)}</a>`)
        .join("");
    document.querySelector("#genreSection").hidden = genres.length === 0;
}

async function setupHome() {
    await sessionReady;

    const [recommended, newest, topRated, genres] = await Promise.all([
        session.user ? api("/me/recommendations").catch(() => []) : [],
        api("/titles?sort=newest&limit=12").catch(() => []),
        api("/titles?sort=top&min_rating=0.5&limit=12").catch(() => []),
        api("/genres").catch(() => [])
    ]);

    fillRow("recommended", recommended);
    fillRow("watchlist", session.savedTitles);
    fillRow("newest", newest);
    fillRow("topRated", topRated);
    renderGenres(genres);
    renderServiceBar();

    const pick = recommended.find((title) => title.poster_url)
        ?? newest.find((title) => title.poster_url)
        ?? recommended[0]
        ?? newest[0];

    if (pick) {
        renderHero(await api(`/titles/${pick.id}`).catch(() => pick));
    }

    if (!pick && !genres.length) {
        const message = document.querySelector("#homeMessage");
        message.textContent = "Couldn't load movies.";
        message.hidden = false;
    }
}

setupHome();
