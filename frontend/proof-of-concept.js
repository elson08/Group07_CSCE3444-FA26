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
function setupMoviePage() {
    const movieHero = document.querySelector("#movieHero");

    // The movie detail elements only exist on movie.html.
    if (!movieHero) {
        return;
    }

    const urlParameters = new URLSearchParams(window.location.search);
    const movie = getMovieById(urlParameters.get("id"));

    document.title = `7Movies | ${movie.title}`;

    movieHero.innerHTML = `
        <img
            class="poster"
            src="${movie.poster}"
            alt="${movie.title} poster"
            onerror="this.src='https://placehold.co/500x750/222c3d/ffffff?text=Movie+Poster'"
        >

        <div>
            <h1>${movie.title}</h1>
            <strong>${movie.year} · ${movie.genre}</strong>
            <p>${movie.description}</p>
        </div>
    `;

    const subscriptionList = document.querySelector("#subscriptionList");
    subscriptionList.innerHTML = movie.subscriptions
        .map((service) => `<li>${service}</li>`)
        .join("");

    const trailerFrame = document.querySelector("#trailer");
    trailerFrame.src = `https://www.youtube.com/embed/${movie.trailer}`;

    // Basic search bar logic, copied from the setUpMoviePage function
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


// Initialize only the functionality relevant to the page being viewed.
setupHomePage();
setupProfilePage();
setupMoviePage();