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

/**
 * setting up a placeholder user
 */
function setupProfilePage() {
    const collections = {
        movies: ["dune2", "arrival", "inception", "matrix", "oppenheimer"],
        shows: ["interstellar", "batman", "spiderverse"],
        subscriptions: []
    };

    Object.entries(collections).forEach(([containerId, movieIds]) => {
        const container = document.getElementById(containerId);

        if (!container) {
            return;
        }

        container.innerHTML = movieIds
            .map((movieId) => {
                const movie = getMovieById(movieId);

                return `
                    <div class="list-entry">
                        <img
                            src="${movie.poster}"
                            alt=""
                            onerror="this.src='https://placehold.co/100x150/222c3d/ffffff?text=Poster'"
                        >

                        <a href="movie.html?id=${movie.id}">
                            ${movie.title}
                        </a>

                        <button
                            type="button"
                            class="delete-movie"
                        >
                            X
                        </button>
                    </div>
                `;
            })
            .join("");
    });
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