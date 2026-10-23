// This javascript is all to show proof of concept for our website, this is not to be used in the final product

// List of some hardcoded movies with the same information we should get from the IMDb data set hopefully
const movies = [
    {
        id: "dune2",
        title: "Dune: Part Two",
        year: 2024,
        genre: "Sci-fi / Adventure",
        description:
            "Paul Atreides unites with Chani and the Fremen while seeking revenge against the conspirators who destroyed his family.",
        poster: "https://image.tmdb.org/t/p/w500/1pdfLvkbY9ohJlCjQH2CZjjYVvJ.jpg",
        trailer: "Way9Dexny3w",
        subscriptions: ["Max", "Prime Video (rent/buy)"]
    },
    {
        id: "interstellar",
        title: "Interstellar",
        year: 2014,
        genre: "Sci-fi / Drama",
        description:
            "A team of explorers travels beyond this galaxy to discover whether mankind has a future among the stars.",
        poster: "https://www.themoviedb.org/t/p/w600_and_h900_face/yQvGrMoipbRoddT0ZR8tPoR7NfX.jpg",
        trailer: "zSWdZVtXT7E",
        subscriptions: ["Paramount+", "Prime Video (rent/buy)"]
    },
    {
        id: "batman",
        title: "The Batman",
        year: 2022,
        genre: "Action / Crime",
        description:
            "Batman investigates corruption in Gotham City while pursuing a mysterious serial killer.",
        poster: "https://image.tmdb.org/t/p/w500/74xTEgt7R36Fpooo50r9T25onhq.jpg",
        trailer: "mqqft2x_Aa4",
        subscriptions: ["Max", "Prime Video (rent/buy)"]
    },
    {
        id: "arrival",
        title: "Arrival",
        year: 2016,
        genre: "Sci-fi / Drama",
        description:
            "A linguist works with the military to communicate with mysterious spacecraft that have landed around the world.",
        poster: "https://www.themoviedb.org/t/p/w600_and_h900_face/pEzNVQfdzYDzVK0XqxERIw2x2se.jpg",
        trailer: "tFMo3UJ4B4g",
        subscriptions: ["Paramount+", "Prime Video (rent/buy)"]
    },
    {
        id: "inception",
        title: "Inception",
        year: 2010,
        genre: "Sci-fi / Thriller",
        description:
            "A skilled thief who steals secrets through dream-sharing is offered a chance to erase his criminal past.",
        poster: "https://image.tmdb.org/t/p/w500/oYuLEt3zVCKq57qu2F8dT7NIa6f.jpg",
        trailer: "YoHD9XEInc0",
        subscriptions: ["Max", "Prime Video (rent/buy)"]
    },
    {
        id: "matrix",
        title: "The Matrix",
        year: 1999,
        genre: "Sci-fi / Action",
        description:
            "A hacker discovers that the reality he knows is a simulated world and joins a rebellion against its creators.",
        poster: "https://image.tmdb.org/t/p/w500/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg",
        trailer: "vKQi3bBA1y8",
        subscriptions: ["Max", "Prime Video (rent/buy)"]
    },
    {
        id: "oppenheimer",
        title: "Oppenheimer",
        year: 2023,
        genre: "Drama / History",
        description:
            "The story of J. Robert Oppenheimer and the scientific and personal consequences of the atomic-bomb project.",
        poster: "",
        trailer: "uYPbbksJxIg",
        subscriptions: ["Peacock", "Prime Video (rent/buy)"]
    },
    {
        id: "spiderverse",
        title: "Spider-Man: Across the Spider-Verse",
        year: 2023,
        genre: "Animation / Action",
        description:
            "Miles Morales journeys across the multiverse and encounters a team of Spider-People.",
        poster: "",
        trailer: "cqGjhVJWtEg",
        subscriptions: ["Netflix", "Prime Video (rent/buy)"]
    }
];

/**
 * attempts to find a movie by its id.
 */
function getMovieById(id) {
    return movies.find((movie) => movie.id === id);
}

/**
 * Create the HTML for a clickable movie card.
 */
function createMovieCard(movie) {
    return `
        <article class="movie-card">
            <a href="movie.html?id=${encodeURIComponent(movie.id)}">
                <img
                    class="poster"
                    src="${movie.poster}"
                    alt="${movie.title} poster"
                    loading="lazy"
                    onerror="this.src='https://placehold.co/500x750/222c3d/ffffff?text=Movie+Poster'"
                >
                <h3>${movie.title}</h3>
            </a>
        </article>
    `;
}

/**
 * Render the Top 3 and Other Recommendations sections.
 * The search field filters both sections by movie title.
 */
function setupHomePage() {

    // DOM accessors so the javascript can edit and access the html
    const topMoviesContainer = document.querySelector("#topMovies");
    const otherMoviesContainer = document.querySelector("#otherMovies");
    const emptyState = document.querySelector("#emptyState");
    const searchInput = document.querySelector("#searchInput");
    const searchForm = document.querySelector("#searchForm");
    
    // If either of elements aren't on the current page, we're not on the home page or something went wrong so we stop here.
    if (!topMoviesContainer || !otherMoviesContainer) {
        return;
    }

    // helper function to create the movie cards based on if they are in the "top 3" or "other" section
    function renderMovies(searchText = "") {
        const normalizedSearch = searchText.trim().toLowerCase();


        const matchingMovies = movies.filter((movie) =>
            movie.title.toLowerCase().includes(normalizedSearch)
        );

        // Splits the topMovies section into the first 3 movies from the movies container and joins them together
        topMoviesContainer.innerHTML = matchingMovies
            .slice(0, 3)
            .map(createMovieCard)
            .join("");

        // Splits the otherMovies section into the rest of the movies from the movies container and joins them together
        otherMoviesContainer.innerHTML = matchingMovies
            .slice(3, 8)
            .map(createMovieCard)
            .join("");

        emptyState.classList.toggle("hidden", matchingMovies.length > 0);
    }

    renderMovies();

    // the basic search bar logic, this needs to be changed to have a dynamic list of possible movie choices for the user
    searchForm.addEventListener("submit", (event) => {
        event.preventDefault();

        const search = searchInput.value.trim().toLowerCase();

        // Won't search if the input is empty.
        if (search === "") {
            return;
        }

        // We first look for an exact match
        let movie = movies.find(
            (movie) => movie.title.toLowerCase() === search
        );

        // If there is no exact match, then we look for a partial match that includes what we typed
        if (!movie) {
            movie = movies.find(
                (movie) => movie.title.toLowerCase().includes(search)
            );
        }

        // If a movie is found, go to its movie page
        if (movie) {
            window.location.href = `movie.html?id=${movie.id}`;
        } else {
            // Show a message if no movie matches.
            emptyState.textContent = "No movies found. Try another title.";
            emptyState.classList.remove("hidden");
        }
    });
}

/**
 * Watchlist and viewing history use localStorage for this project prototype.
 * A future backend integration can replace these helpers with /api/me/titles.
 */
const WATCHLIST_STORAGE_KEY = "7movies.watchlist.v1";
const HISTORY_STORAGE_KEY = "7movies.viewingHistory.v1";

function readStoredIds(key) {
    try {
        const value = JSON.parse(localStorage.getItem(key) || "[]");
        return Array.isArray(value) ? value.filter((id) => typeof id === "string") : [];
    } catch (error) {
        console.warn(`Could not read ${key} from browser storage.`, error);
        return [];
    }
}

function writeStoredIds(key, ids) {
    localStorage.setItem(key, JSON.stringify([...new Set(ids)]));
}

function addStoredId(key, id) {
    const ids = readStoredIds(key);
    if (!ids.includes(id)) ids.unshift(id);
    writeStoredIds(key, ids);
}

function removeStoredId(key, id) {
    writeStoredIds(key, readStoredIds(key).filter((storedId) => storedId !== id));
}

function renderSavedList(container, ids, emptyMessage, options = {}) {
    if (!container) return;
    const { allowRemove = false, allowMarkWatched = false } = options;
    const validMovies = ids.map(getMovieById).filter(Boolean);

    if (!validMovies.length) {
        container.innerHTML = `<p class="empty-list">${emptyMessage}</p>`;
        return;
    }

    container.innerHTML = validMovies.map((movie) => `
        <article class="list-entry">
            <img src="${movie.poster || "https://placehold.co/100x150/222c3d/ffffff?text=Poster"}"
                 alt="${movie.title} poster" loading="lazy">
            <div class="saved-title-info">
                <a href="movie.html?id=${encodeURIComponent(movie.id)}">${movie.title}</a>
                <small>${movie.year} · ${movie.genre}</small>
                ${allowMarkWatched ? `<button type="button" class="small-action" data-mark-watched="${movie.id}">Mark watched</button>` : ""}
            </div>
            ${allowRemove ? `<button type="button" class="remove-saved" data-remove-id="${movie.id}" aria-label="Remove ${movie.title}">Remove</button>` : ""}
        </article>
    `).join("");

    container.querySelectorAll("[data-remove-id]").forEach((button) => {
        button.addEventListener("click", () => {
            removeStoredId(WATCHLIST_STORAGE_KEY, button.dataset.removeId);
            setupProfilePage();
        });
    });

    container.querySelectorAll("[data-mark-watched]").forEach((button) => {
        button.addEventListener("click", () => {
            const id = button.dataset.markWatched;
            addStoredId(HISTORY_STORAGE_KEY, id);
            removeStoredId(WATCHLIST_STORAGE_KEY, id);
            setupProfilePage();
        });
    });
}

/**
 * Profile collections: movies, shows, and streaming subscriptions.
 * This prototype saves the collections in localStorage, so they persist in this browser.
 * The demo catalog currently uses movie titles as suggestions for the Shows collection too.
 */
function setupProfileCollections() {
    const panels = [
        { key: "movies", inputId: "movieSearch", options: movies },
        { key: "shows", inputId: "showSearch", options: movies },
        { key: "subscriptions", inputId: "subscriptionSearch", options: [...new Set(movies.flatMap(movie => movie.subscriptions))].map(title => ({ id: title, title })) }
    ];
    if (!document.querySelector("#movies")) return;

    const storageKey = "7movies-demo-collections";
    let saved = {};
    try {
        saved = JSON.parse(localStorage.getItem(storageKey) || "{}");
        if (!saved || typeof saved !== "object" || Array.isArray(saved)) saved = {};
    } catch (error) {
        console.warn("Could not read saved profile collections.", error);
    }
    const collections = Object.fromEntries(panels.map(panel => [
        panel.key,
        Array.isArray(saved[panel.key]) ? saved[panel.key].filter(value => typeof value === "string") : []
    ]));
    const persist = () => {
        try {
            localStorage.setItem(storageKey, JSON.stringify(collections));
        } catch (error) {
            console.error("Could not save profile collections.", error);
        }
    };

    panels.forEach(panel => {
        const container = document.getElementById(panel.key);
        const input = document.getElementById(panel.inputId);
        if (!container || !input) return;
        const addButton = input.parentElement.querySelector("button");

        // Native datalist suggestions appear as the user types.
        const datalist = document.createElement("datalist");
        datalist.id = `${panel.inputId}-suggestions`;
        panel.options.forEach(item => {
            const option = document.createElement("option");
            option.value = item.title;
            datalist.append(option);
        });
        input.setAttribute("list", datalist.id);
        input.after(datalist);

        const message = document.createElement("p");
        message.className = "form-message";
        message.setAttribute("aria-live", "polite");
        input.parentElement.after(message);

        function render() {
            container.replaceChildren();
            if (collections[panel.key].length === 0) {
                const empty = document.createElement("p");
                empty.className = "empty-list";
                empty.textContent = "Nothing added yet.";
                container.append(empty);
                return;
            }
            collections[panel.key].forEach(id => {
                const item = panel.options.find(option => option.id === id);
                if (!item) return;
                const row = document.createElement("div");
                row.className = "list-entry";
                const label = document.createElement("span");
                label.textContent = item.title;
                label.style.flex = "1";
                const removeButton = document.createElement("button");
                removeButton.type = "button";
                removeButton.className = "delete-movie";
                removeButton.textContent = "Remove";
                removeButton.setAttribute("aria-label", `Remove ${item.title}`);
                removeButton.addEventListener("click", () => {
                    collections[panel.key] = collections[panel.key].filter(value => value !== id);
                    persist();
                    message.textContent = `${item.title} removed.`;
                    render();
                });
                row.append(label, removeButton);
                container.append(row);
            });
        }

        function addItem() {
            const query = input.value.trim().toLowerCase();
            if (!query) {
                message.textContent = "Type a title or subscription before adding.";
                input.focus();
                return;
            }
            const item = panel.options.find(option => option.title.toLowerCase() === query);
            if (!item) {
                message.textContent = "Choose an item from the suggestions list.";
                return;
            }
            if (collections[panel.key].includes(item.id)) {
                message.textContent = `${item.title} is already in your list.`;
                return;
            }
            collections[panel.key].push(item.id);
            persist();
            input.value = "";
            message.textContent = `${item.title} added.`;
            render();
        }

        addButton?.addEventListener("click", addItem);
        input.addEventListener("keydown", event => {
            if (event.key === "Enter") {
                event.preventDefault();
                addItem();
            }
        });
        render();
    });
}

/** Render the user's watchlist and viewing history on profile.html. */
function setupProfilePage() {
    const watchlist = document.querySelector("#watchlist");
    const history = document.querySelector("#viewingHistory");
    if (!watchlist && !history) return;

    renderSavedList(watchlist, readStoredIds(WATCHLIST_STORAGE_KEY),
        "Your watchlist is empty. Open a movie and choose Add to Watchlist.",
        { allowRemove: true, allowMarkWatched: true });
    renderSavedList(history, readStoredIds(HISTORY_STORAGE_KEY),
        "Your viewing history is empty. Mark a title as watched to see it here.",
        { allowRemove: false });
}

/** Fill the movie detail page and connect its watchlist controls. */
function setupMoviePage() {
    const movieHero = document.querySelector("#movieHero");
    if (!movieHero) return;

    const urlParameters = new URLSearchParams(window.location.search);
    const movie = getMovieById(urlParameters.get("id"));
    if (!movie) {
        movieHero.innerHTML = "<h1>Movie not found</h1><p>Return to the homepage and choose a title.</p>";
        const actions = document.querySelector(".movie-actions");
        if (actions) actions.hidden = true;
        return;
    }

    document.title = `7Movies | ${movie.title}`;
    movieHero.innerHTML = `
        <img class="poster"
             src="${movie.poster || "https://placehold.co/500x750/222c3d/ffffff?text=Movie+Poster"}"
             alt="${movie.title} poster">
        <div>
            <h1>${movie.title}</h1>
            <strong>${movie.year} · ${movie.genre}</strong>
            <p>${movie.description}</p>
        </div>
    `;

    const subscriptionList = document.querySelector("#subscriptionList");
    if (subscriptionList) {
        subscriptionList.innerHTML = movie.subscriptions.map((service) => `<li>${service}</li>`).join("");
    }

    const trailerFrame = document.querySelector("#trailer");
    if (trailerFrame && movie.trailer) trailerFrame.src = `https://www.youtube.com/embed/${movie.trailer}`;

    const watchlistButton = document.querySelector("#watchlistButton");
    const watchedButton = document.querySelector("#watchedButton");
    const message = document.querySelector("#watchlistMessage");

    // Opening a title records it in viewing history, unless it is already there.
    addStoredId(HISTORY_STORAGE_KEY, movie.id);

    function updateActionLabels() {
        const inWatchlist = readStoredIds(WATCHLIST_STORAGE_KEY).includes(movie.id);
        const watched = readStoredIds(HISTORY_STORAGE_KEY).includes(movie.id);
        if (watchlistButton) watchlistButton.textContent = inWatchlist ? "✓ In Watchlist (Remove)" : "＋ Add to Watchlist";
        if (watchedButton) watchedButton.textContent = watched ? "✓ Watched (Mark Unwatched)" : "✓ Mark as Watched";
    }

    watchlistButton?.addEventListener("click", () => {
        if (readStoredIds(WATCHLIST_STORAGE_KEY).includes(movie.id)) {
            removeStoredId(WATCHLIST_STORAGE_KEY, movie.id);
            if (message) message.textContent = `${movie.title} removed from your watchlist.`;
        } else {
            addStoredId(WATCHLIST_STORAGE_KEY, movie.id);
            if (message) message.textContent = `${movie.title} added to your watchlist.`;
        }
        updateActionLabels();
    });

    watchedButton?.addEventListener("click", () => {
        if (readStoredIds(HISTORY_STORAGE_KEY).includes(movie.id)) {
            removeStoredId(HISTORY_STORAGE_KEY, movie.id);
            if (message) message.textContent = `${movie.title} removed from viewing history.`;
        } else {
            addStoredId(HISTORY_STORAGE_KEY, movie.id);
            removeStoredId(WATCHLIST_STORAGE_KEY, movie.id);
            if (message) message.textContent = `${movie.title} marked as watched.`;
        }
        updateActionLabels();
    });

    updateActionLabels();
}

// Initialize only the functionality relevant to the page being viewed.
setupHomePage();
setupProfilePage();
setupProfileCollections();
setupMoviePage();