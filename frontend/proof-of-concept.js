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

function escapeHTML(value) {
    const element = document.createElement("div");
    element.textContent = value;
    return element.innerHTML.replaceAll('"', "&quot;");
}

async function fetchAPI(path) {
    const response = await fetch(`/api${path}`);
    if (!response.ok) {
        throw new Error(`Request failed: ${response.status}`);
    }
    return response.json();
}

function titleToMovie(title) {
    return {
        id: title.id,
        title: title.name,
        year: title.release_year ?? "",
        genre: (title.genres ?? []).map((genre) => genre.name).join(" / "),
        description: title.synopsis || "No description available.",
        poster: title.poster_url ?? "",
        trailerUrl: title.trailer_url,
        subscriptions: (title.services ?? []).map((service) => service.name)
    };
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
                    src="${escapeHTML(movie.poster)}"
                    alt="${escapeHTML(movie.title)} poster"
                    loading="lazy"
                    onerror="this.src='https://placehold.co/500x750/222c3d/ffffff?text=Movie+Poster'"
                >
                <h3>${escapeHTML(movie.title)}</h3>
            </a>
        </article>
    `;
}

/**
 * Render the Top 3 and Other Recommendations sections.
 * The search field filters both sections by movie title.
 */
async function setupHomePage() {

    // DOM accessors so the javascript can edit and access the html
    const topMoviesContainer = document.querySelector("#topMovies");
    const otherMoviesContainer = document.querySelector("#otherMovies");
    const emptyState = document.querySelector("#emptyState");
    const searchInput = document.querySelector("#searchInput");

    // If either of elements aren't on the current page, we're not on the home page or something went wrong so we stop here.
    if (!topMoviesContainer || !otherMoviesContainer) {
        return;
    }

    const searchText = new URLSearchParams(window.location.search).get("q") ?? "";
    searchInput.value = searchText;

    let matchingMovies = [];
    try {
        let titles = [];
        if (!searchText) {
            titles = await fetchAPI("/me/recommendations").catch(() => []);
        }

        if (titles.length === 0) {
            const query = new URLSearchParams({ q: searchText, limit: searchText ? 50 : 8 });
            titles = await fetchAPI(`/titles?${query}`);
        }

        matchingMovies = titles.map(titleToMovie);
        emptyState.textContent = "No movies found. Try another title.";
    } catch {
        emptyState.textContent = "Could not load movies. Please try again later.";
    }

    // Splits the topMovies section into the first 3 movies from the movies container and joins them together
    topMoviesContainer.innerHTML = matchingMovies
        .slice(0, 3)
        .map(createMovieCard)
        .join("");

    // Splits the otherMovies section into the rest of the movies from the movies container and joins them together
    otherMoviesContainer.innerHTML = matchingMovies
        .slice(3)
        .map(createMovieCard)
        .join("");

    emptyState.classList.toggle("hidden", matchingMovies.length > 0);
}

function setupSearch() {
    const searchForm = document.querySelector("#searchForm");

    if (!searchForm) {
        return;
    }

    searchForm.addEventListener("submit", (event) => {
        event.preventDefault();

        const search = document.querySelector("#searchInput").value.trim();

        // Won't search if the input is empty.
        if (search !== "") {
            window.location.href = `index.html?q=${encodeURIComponent(search)}`;
        }
    });
}

// /**
//  * setting up a placeholder user
//  */
// function setupProfilePage() {
//     const collections = {
//         movies: ["dune2", "arrival", "inception", "matrix", "oppenheimer"],
//         shows: ["interstellar", "batman", "spiderverse"],
//         subscriptions: []
//     };

//     Object.entries(collections).forEach(([containerId, movieIds]) => {
//         const container = document.getElementById(containerId);

//         if (!container) {
//             return;
//         }

//         container.innerHTML = movieIds
//             .map((movieId) => {
//                 const movie = getMovieById(movieId);

//                 return `
//                     <div class="list-entry">
//                         <img
//                             src="${movie.poster}"
//                             alt=""
//                             onerror="this.src='https://placehold.co/100x150/222c3d/ffffff?text=Poster'"
//                         >

//                         <a href="movie.html?id=${movie.id}">
//                             ${movie.title}
//                         </a>

//                         <button
//                             type="button"
//                             class="delete-movie"
//                         >
//                             X
//                         </button>
//                     </div>
//                 `;
//             })
//             .join("");
//     });
// }

/**
 * Send requests for the signed-in user's profile collections.
 * The existing backend routes use the session cookie to identify the user.
 */
async function profileRequest(path, method = "GET") {
    const response = await fetch(`/api${path}`, {
        method,
        credentials: "same-origin",
        cache: "no-store"
    });

    if (response.status === 204) {
        return null;
    }

    const contentType = response.headers.get("Content-Type") || "";
    const data = contentType.includes("application/json")
        ? await response.json()
        : null;

    if (!response.ok) {
        throw new Error(data?.error || `Request failed (${response.status}).`);
    }

    return data;
}

// Display one title from the database in the Movies or Shows collection.
function createProfileTitleEntry(title) {
    const poster = title.poster_url ||
        "https://placehold.co/100x150/222c3d/ffffff?text=Poster";
    const year = title.release_year ? ` · ${title.release_year}` : "";

    return `
        <article class="list-entry">
            <img src="${escapeHTML(poster)}"
                 alt="${escapeHTML(title.name)} poster"
                 loading="lazy"
                 onerror="this.src='https://placehold.co/100x150/222c3d/ffffff?text=Poster'">
            <a href="movie.html?id=${encodeURIComponent(title.id)}">
                ${escapeHTML(title.name)}${year}
            </a>
            <button type="button" class="delete-movie"
                    data-remove-title="${title.id}"
                    aria-label="Remove ${escapeHTML(title.name)}">
                X
            </button>
        </article>
    `;
}

/** Display one streaming service from the database. */
function createProfileServiceEntry(service) {
    return `
        <article class="list-entry service-entry">
            <span>${escapeHTML(service.name)}</span>
            <button type="button" class="delete-movie"
                    data-remove-service="${service.id}"
                    aria-label="Remove ${escapeHTML(service.name)}">
                Remove
            </button>
        </article>
    `;
}

/**
 * Connect one profile search box to the existing catalog API.
 * The user must select an actual database result before pressing Add.
 */
function setupProfileSearch({ inputId, kind, addPath, searchPath, onAdded }) {
    const input = document.getElementById(inputId);
    if (!input) return;

    const addButton = input.parentElement.querySelector("button");
    const results = document.createElement("div");
    results.className = "profile-search-results";
    results.hidden = true;
    input.parentElement.insertAdjacentElement("afterend", results);

    const message = document.createElement("p");
    message.className = "form-message profile-list-message";
    message.setAttribute("aria-live", "polite");
    results.insertAdjacentElement("afterend", message);

    let matches = [];
    let selected = null;
    let requestNumber = 0;

    function clearSelection() {
        selected = null;
        if (addButton) addButton.disabled = true;
    }

    function showResults(items, labelForItem) {
        matches = items;
        results.replaceChildren();
        results.hidden = items.length === 0;

        items.forEach((item, index) => {
            const option = document.createElement("button");
            option.type = "button";
            option.className = "profile-search-option";
            option.textContent = labelForItem(item);
            option.addEventListener("click", () => {
                selected = matches[index];
                input.value = labelForItem(selected);
                results.hidden = true;
                message.textContent = `Selected ${labelForItem(selected)}.`;
                if (addButton) addButton.disabled = false;
            });
            results.append(option);
        });
    }

    input.addEventListener("input", async () => {
        const query = input.value.trim();
        const currentRequest = ++requestNumber;
        clearSelection();
        message.textContent = "";

        if (!query) {
            results.hidden = true;
            results.replaceChildren();
            return;
        }

        try {
            let items;
            if (kind === "service") {
                const allServices = await profileRequest("/services");
                items = allServices.filter(service =>
                    service.name.toLowerCase().includes(query.toLowerCase())
                ).slice(0, 10);
            } else {
                const params = new URLSearchParams({ q: query, kind, limit: "10" });
                items = await profileRequest(`/titles?${params.toString()}`);
            }

            // Ignore an older search if the user has already typed something else.
            if (currentRequest !== requestNumber) return;

            showResults(items, item => kind === "service"
                ? item.name
                : `${item.name}${item.release_year ? ` (${item.release_year})` : ""}`);

            if (items.length === 0) {
                message.textContent = "No matching results found.";
            }
        } catch (error) {
            if (currentRequest === requestNumber) {
                results.hidden = true;
                message.textContent = error.message;
            }
        }
    });

    addButton?.addEventListener("click", async () => {
        if (!selected) {
            message.textContent = "Search and select an item before adding it.";
            return;
        }

        const item = selected;
        addButton.disabled = true;
        try {
            await profileRequest(`${addPath}/${item.id}`, "PUT");
            input.value = "";
            selected = null;
            results.hidden = true;
            results.replaceChildren();
            message.textContent = `${item.name} added to your profile.`;
            await onAdded();
        } catch (error) {
            message.textContent = error.message;
        } finally {
            addButton.disabled = !selected;
        }
    });

    input.addEventListener("keydown", event => {
        if (event.key === "Enter") {
            event.preventDefault();
            if (selected) addButton?.click();
            else message.textContent = "Choose an item from the search results first.";
        }
    });

    if (addButton) addButton.disabled = true;
}

/** Load and render all three database-backed profile collections. */
async function setupProfilePage() {
    const moviesContainer = document.getElementById("movies");
    const showsContainer = document.getElementById("shows");
    const subscriptionsContainer = document.getElementById("subscriptions");

    // This script is also included on pages other than profile.html.
    if (!moviesContainer || !showsContainer || !subscriptionsContainer) return;

    const pageMessage = document.getElementById("profileCollectionsMessage");

    async function loadCollections() {
        if (pageMessage) {
            pageMessage.textContent = "Loading your collections...";
        }

        try {
            const [titlesResponse, servicesResponse] = await Promise.all([
                profileRequest("/me/titles"),
                profileRequest("/me/services")
            ]);

            // Make sure empty database responses are treated as empty arrays.
            const savedTitles = Array.isArray(titlesResponse)
                ? titlesResponse
                : [];

            const savedServices = Array.isArray(servicesResponse)
                ? servicesResponse
                : [];

            // Separate movies and shows.
            const savedMovies = savedTitles.filter(
                title => title.kind === "movie"
            );

            const savedShows = savedTitles.filter(
                title => title.kind === "show"
            );

            // Display movies or an empty-state message.
            moviesContainer.innerHTML = savedMovies.length
                ? savedMovies.map(createProfileTitleEntry).join("")
                : `
                    <div class="empty-list">
                        <span class="empty-list-icon" aria-hidden="true">▣</span>
                        <strong>Your movie collection is empty</strong>
                        <span>Search for a movie above to start building your collection.</span>
                    </div>
                `;

            // Display shows or an empty-state message.
            showsContainer.innerHTML = savedShows.length
                ? savedShows.map(createProfileTitleEntry).join("")
                : `
                    <div class="empty-list">
                        <span class="empty-list-icon" aria-hidden="true">▤</span>
                        <strong>No shows added yet</strong>
                        <span>Find your favorite TV shows and keep them here.</span>
                    </div>
                `;

            // Display subscriptions or an empty-state message.
            subscriptionsContainer.innerHTML = savedServices.length
                ? savedServices.map(createProfileServiceEntry).join("")
                : `
                    <div class="empty-list">
                        <span class="empty-list-icon" aria-hidden="true">☆</span>
                        <strong>No subscriptions added</strong>
                        <span>Add the streaming services you use to keep them together.</span>
                    </div>
                `;

            // Remove movie buttons.
            moviesContainer.querySelectorAll("[data-remove-title]").forEach(button => {
                button.addEventListener("click", async () => {
                    button.disabled = true;

                    try {
                        await profileRequest(
                            `/me/titles/${button.dataset.removeTitle}`,
                            "DELETE"
                        );

                        await loadCollections();
                    } catch (error) {
                        if (pageMessage) {
                            pageMessage.textContent = error.message;
                        }

                        button.disabled = false;
                    }
                });
            });

            // Remove show buttons.
            showsContainer.querySelectorAll("[data-remove-title]").forEach(button => {
                button.addEventListener("click", async () => {
                    button.disabled = true;

                    try {
                        await profileRequest(
                            `/me/titles/${button.dataset.removeTitle}`,
                            "DELETE"
                        );

                        await loadCollections();
                    } catch (error) {
                        if (pageMessage) {
                            pageMessage.textContent = error.message;
                        }

                        button.disabled = false;
                    }
                });
            });

            // Remove subscription buttons.
            subscriptionsContainer.querySelectorAll("[data-remove-service]").forEach(button => {
                button.addEventListener("click", async () => {
                    button.disabled = true;

                    try {
                        await profileRequest(
                            `/me/services/${button.dataset.removeService}`,
                            "DELETE"
                        );

                        await loadCollections();
                    } catch (error) {
                        if (pageMessage) {
                            pageMessage.textContent = error.message;
                        }

                        button.disabled = false;
                    }
                });
            });

            if (pageMessage) {
                pageMessage.textContent = "";
            }

        } catch (error) {
            if (pageMessage) {
                pageMessage.textContent =
                    `${error.message} Make sure you are signed in, then refresh the page.`;
            }
        }
    }

    // These routes already exist in the Go backend; no new backend endpoints are needed.
    setupProfileSearch({
        inputId: "movieSearch",
        kind: "movie",
        addPath: "/me/titles",
        onAdded: loadCollections
    });

    setupProfileSearch({
        inputId: "showSearch",
        kind: "show",
        addPath: "/me/titles",
        onAdded: loadCollections
    });

    setupProfileSearch({
        inputId: "subscriptionSearch",
        kind: "service",
        addPath: "/me/services",
        onAdded: loadCollections
    });

    await loadCollections();
}

/**
 * Fill the movie detail page using the movie ID in the URL.
 */
async function setupMoviePage() {
    const movieHero = document.querySelector("#movieHero");

    // The movie detail elements only exist on movie.html.
    if (!movieHero) {
        return;
    }

    const id = new URLSearchParams(window.location.search).get("id");

    let movie = getMovieById(id);
    const fromAPI = !movie;
    if (fromAPI) {
        try {
            movie = titleToMovie(await fetchAPI(`/titles/${encodeURIComponent(id)}`));
        } catch {
            movieHero.innerHTML = "<h1>Movie not found.</h1>";
            return;
        }
    }

    document.title = `7Movies | ${movie.title}`;

    movieHero.innerHTML = `
        <img
            class="poster"
            src="${escapeHTML(movie.poster)}"
            alt="${escapeHTML(movie.title)} poster"
            onerror="this.src='https://placehold.co/500x750/222c3d/ffffff?text=Movie+Poster'"
        >

        <div>
            <h1>${escapeHTML(movie.title)}</h1>
            <strong>${escapeHTML(movie.year)} · ${escapeHTML(movie.genre)}</strong>
            <p>${escapeHTML(movie.description)}</p>
            ${fromAPI ? '<div class="add-to-profile"><button id="saveButton" type="button" disabled>Save to profile</button></div>' : ""}
        </div>
    `;

    if (fromAPI) {
        setupSaveButton(id);
    }

    const subscriptionList = document.querySelector("#subscriptionList");
    subscriptionList.innerHTML = movie.subscriptions
        .map((service) => `<li>${escapeHTML(service)}</li>`)
        .join("");

    const trailerSrc = movie.trailerUrl ?? (movie.trailer && `https://www.youtube.com/embed/${movie.trailer}`);
    if (trailerSrc) {
        document.querySelector("#trailer").src = trailerSrc;
    }
}

async function setupSaveButton(id) {
    const button = document.querySelector("#saveButton");

    let saved;
    try {
        const savedTitles = await fetchAPI("/me/titles");
        saved = savedTitles.some((title) => String(title.id) === id);
    } catch {
        button.textContent = "Sign in to save";
        button.disabled = false;
        button.addEventListener("click", () => window.location.assign("signin.html"));
        return;
    }

    const showState = () => {
        button.textContent = saved ? "Remove from profile" : "Save to profile";
    };

    showState();
    button.disabled = false;

    button.addEventListener("click", async () => {
        button.disabled = true;
        try {
            const response = await fetch(`/api/me/titles/${encodeURIComponent(id)}`, { method: saved ? "DELETE" : "PUT" });
            if (response.ok) {
                saved = !saved;
                showState();
            }
        } finally {
            button.disabled = false;
        }
    });
}


// Initialize only the functionality relevant to the page being viewed.
setupSearch();
setupHomePage();
setupProfilePage();
setupMoviePage();