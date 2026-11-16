const pageSize = 24;

function minRatingText(value) {
    return Number(value) > 0 ? `★ ${value} and up` : "Any rating";
}

function filterLabels(params, services) {
    const serviceNames = new Map(services.map((service) => [String(service.id), service.name]));
    const labels = {
        q: (value) => `“${value}”`,
        kind: (value) => (value === "show" ? "Shows" : "Movies"),
        genre: (value) => value,
        service: (value) => serviceNames.get(value) ?? "Service",
        mine: () => "My services",
        year_from: (value) => `From ${value}`,
        year_to: (value) => `To ${value}`,
        min_rating: (value) => `★ ${value}+`
    };

    return [...params]
        .filter(([key]) => labels[key])
        .map(([key, value]) => {
            const remaining = new URLSearchParams([...params].filter(([k, v]) => !(k === key && v === value) && k !== "page"));
            return `<a class="chip active" href="?${remaining}">${escapeHTML(labels[key](value))} ×</a>`;
        })
        .join("");
}

function renderPagination(page, pages, params) {
    const nav = document.querySelector("#pagination");
    if (pages <= 1) {
        nav.innerHTML = "";
        return;
    }

    const href = (number) => {
        const next = new URLSearchParams(params);
        next.set("page", number);
        return `?${next}`;
    };

    const parts = [];
    if (page > 1) {
        parts.push(`<a href="${href(page - 1)}">‹ Prev</a>`);
    }

    const numbers = [...new Set([1, page - 2, page - 1, page, page + 1, page + 2, pages])]
        .filter((number) => number >= 1 && number <= pages)
        .sort((a, b) => a - b);

    let previous = 0;
    numbers.forEach((number) => {
        if (number - previous > 1) {
            parts.push("<span>…</span>");
        }
        parts.push(number === page
            ? `<a class="current" aria-current="page" href="${href(number)}">${number}</a>`
            : `<a href="${href(number)}">${number}</a>`);
        previous = number;
    });

    if (page < pages) {
        parts.push(`<a href="${href(page + 1)}">Next ›</a>`);
    }

    nav.innerHTML = parts.join("");
}

async function loadResults(params, services) {
    const page = Math.max(1, Number(params.get("page")) || 1);
    const query = new URLSearchParams(params);
    query.delete("page");
    query.set("limit", pageSize);
    query.set("offset", (page - 1) * pageSize);

    const grid = document.querySelector("#results");
    const summary = document.querySelector("#resultSummary");
    grid.innerHTML = `<p class="no-results">Loading titles...</p>`;
    document.querySelector("#activeFilters").innerHTML = filterLabels(params, services);

    try {
        const { data, total } = await request(`/titles?${query}`);
        renderCards(grid, data, "No titles match these filters. Try removing a few.");
        const search = params.get("q");
        summary.innerHTML = `<strong>${total.toLocaleString()} result${total === 1 ? "" : "s"}</strong>${search ? ` <span class="extraText">for “${escapeHTML(search)}”</span>` : ""}`;
        renderPagination(page, Math.ceil(total / pageSize), params);
    } catch (error) {
        summary.textContent = "";
        grid.innerHTML = `<p class="no-results">${escapeHTML(error.status === 401 ? "Sign in to filter by your services." : error.message)}</p>`;
        document.querySelector("#pagination").innerHTML = "";
    }
}

function fillForm(form, params) {
    form.elements.q.value = params.get("q") ?? "";
    document.querySelector("#headerSearch").value = params.get("q") ?? "";

    const kind = form.querySelector(`input[name="kind"][value="${params.get("kind") ?? ""}"]`);
    if (kind) {
        kind.checked = true;
    }

    form.elements.year_from.value = params.get("year_from") ?? "";
    form.elements.year_to.value = params.get("year_to") ?? "";
    form.elements.min_rating.value = params.get("min_rating") ?? "0";
    document.querySelector("#minRatingLabel").textContent = minRatingText(form.elements.min_rating.value);
    form.elements.mine.checked = params.get("mine") === "1";
    form.elements.sort.value = params.get("sort") ?? "newest";
}

async function setupBrowse() {
    const params = new URLSearchParams(window.location.search);
    const form = document.querySelector("#filters");

    const [genres, services] = await Promise.all([
        api("/genres").catch(() => []),
        api("/services").catch(() => []),
        sessionReady
    ]);

    const chosenGenres = params.getAll("genre");
    document.querySelector("#genreOptions").innerHTML = genres
        .map((genre) => `<label class="chip-check"><input type="checkbox" name="genre" value="${escapeHTML(genre.name)}" ${chosenGenres.includes(genre.name) ? "checked" : ""}><span>${escapeHTML(genre.name)}</span></label>`)
        .join("");

    const chosenServices = params.getAll("service");
    const mine = new Set(session.services.map((service) => service.id));
    document.querySelector("#serviceOptions").innerHTML = [...services]
        .sort((a, b) => (mine.has(b.id) - mine.has(a.id)) || a.name.localeCompare(b.name))
        .map((service) => `<label class="check"><input type="checkbox" name="service" value="${service.id}" ${chosenServices.includes(String(service.id)) ? "checked" : ""}> ${escapeHTML(service.name)}</label>`)
        .join("");

    document.querySelector("#mineOption").hidden = !session.user;
    fillForm(form, params);

    const collapsibleFilters = [
        "Streaming",
        "Genres",
        "Release year",
        "Minimum 7Movies rating"
    ];

    form.querySelectorAll(".filter-group").forEach((group) => {
        const heading = group.querySelector("h3");

        if (!heading || !collapsibleFilters.includes(heading.textContent.trim())) {
            return;
            }

        heading.classList.add("filter-toggle");
        heading.setAttribute("role", "button");
        heading.setAttribute("tabindex", "0");
        heading.setAttribute("aria-expanded", "false");

        group.classList.add("is-collapsed");

        function toggleFilter() {
            const collapsed = group.classList.toggle("is-collapsed");
            heading.setAttribute("aria-expanded", String(!collapsed));
            }

        heading.addEventListener("click", toggleFilter);

        heading.addEventListener("keydown", (event) => {
            if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                toggleFilter();
                }
            });
        });

    form.elements.min_rating.addEventListener("input", () => {
        document.querySelector("#minRatingLabel").textContent = minRatingText(form.elements.min_rating.value);
    });

    form.elements.sort.addEventListener("change", () => form.requestSubmit());

    form.addEventListener("submit", (event) => {
        event.preventDefault();
        const next = new URLSearchParams();
        for (const [key, value] of new FormData(form)) {
            const skip = value === ""
                || (key === "min_rating" && Number(value) === 0)
                || (key === "sort" && value === "newest");
            if (!skip) {
                next.append(key, value);
            }
        }
        window.location.search = next.toString();
    });

    request("/titles?limit=1")
        .then(({ total }) => {
            document.querySelector("#catalogSize").textContent = `${total.toLocaleString()} movies and shows`;
        })
        .catch(() => {});

    await loadResults(params, services);
}

setupBrowse();
