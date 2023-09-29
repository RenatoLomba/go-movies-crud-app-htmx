package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/exp/slices"
)

type Movie struct {
	Id       string
	Isbn     string
	Title    string
	Synopsis string
	Director *Director
}

type Director struct {
	Firstname string
	Lastname  string
}

var movies []Movie

func getMovies(w http.ResponseWriter, r *http.Request) {
	data := map[string][]Movie{
		"Movies": movies,
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html", "templates/movies.html"))
	tmpl.ExecuteTemplate(w, "index.html", data)
}

func newMovie(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Hx-Request") == "true" {
		data := map[string]any{
			"Method":    "post",
			"HxUrl":     "/create-movie",
			"PageTitle": "New Movie",
			"DataForm": &Movie{
				Isbn:     "",
				Title:    "",
				Synopsis: "",
				Director: &Director{
					Firstname: "",
					Lastname:  "",
				},
			},
		}

		tmpl := template.Must(template.ParseFiles("templates/form.html"))
		tmpl.Execute(w, data)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func editMovie(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Hx-Request") == "true" {
		movieId := mux.Vars(r)["id"]

		movieIdx := slices.IndexFunc(movies, func(c Movie) bool { return c.Id == movieId })

		if movieIdx == -1 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		movie := movies[movieIdx]

		data := map[string]any{
			"Method":    "put",
			"HxUrl":     "/update-movie/" + movieId,
			"PageTitle": "Edit Movie",
			"DataForm": &Movie{
				Isbn:     movie.Isbn,
				Title:    movie.Title,
				Synopsis: movie.Synopsis,
				Director: &Director{
					Firstname: movie.Director.Firstname,
					Lastname:  movie.Director.Lastname,
				},
			},
		}

		tmpl := template.Must(template.ParseFiles("templates/form.html"))
		tmpl.Execute(w, data)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	movieId := mux.Vars(r)["id"]

	movieIdx := slices.IndexFunc(movies, func(c Movie) bool { return c.Id == movieId })
	movie := movies[movieIdx]

	data := map[string]Movie{
		"Movie": movie,
	}

	tmpl := template.Must(template.ParseFiles("templates/movie.html"))
	tmpl.Execute(w, data)
}

func deleteMovie(w http.ResponseWriter, r *http.Request) {
	movieId := mux.Vars(r)["id"]

	for idx, movie := range movies {
		if movie.Id == movieId {
			movies = append(movies[:idx], movies[idx+1:]...)
			break
		}
	}

	data := map[string][]Movie{
		"Movies": movies,
	}

	tmpl := template.Must(template.ParseFiles("templates/movies.html"))
	tmpl.Execute(w, data)
}

func updateMovie(w http.ResponseWriter, r *http.Request) {
	movieId := mux.Vars(r)["id"]

	movieIdx := slices.IndexFunc(movies, func(c Movie) bool { return c.Id == movieId })
	movie := movies[movieIdx]

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error when trying to parse form", http.StatusBadRequest)
		return
	}

	movie.Isbn = r.FormValue("isbn")
	movie.Title = r.FormValue("title")
	movie.Synopsis = r.FormValue("synopsis")
	movie.Director.Firstname = r.FormValue("directorfirstname")
	movie.Director.Lastname = r.FormValue("directorlastname")

	movies[movieIdx] = movie

	w.Header().Set("HX-Redirect", "/")
	fmt.Fprintf(w, "Movie updated")
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		http.Error(w, "Error when trying to parse form", http.StatusBadRequest)
		return
	}

	lastMovieId, _ := strconv.Atoi(movies[len(movies)-1].Id)

	movie := Movie{
		Id:       strconv.Itoa(lastMovieId + 1),
		Isbn:     r.FormValue("isbn"),
		Title:    r.FormValue("title"),
		Synopsis: r.FormValue("synopsis"),
		Director: &Director{
			Firstname: r.FormValue("directorfirstname"),
			Lastname:  r.FormValue("directorlastname"),
		},
	}

	movies = append(movies, movie)

	w.Header().Set("HX-Redirect", "/")
	fmt.Fprintf(w, "Movie created")
}

func main() {
	r := mux.NewRouter()

	movies = append(movies, Movie{
		Id:       "1",
		Isbn:     "438227",
		Title:    "Blade Runner",
		Synopsis: "In the twenty-first century, a corporation develops androids to be used as slaves in colonies outside of the Earth...",
		Director: &Director{
			Firstname: "Ridley",
			Lastname:  "Scott",
		},
	})
	movies = append(movies, Movie{
		Id:       "2",
		Isbn:     "454555",
		Title:    "Alien",
		Synopsis: "After a space merchant ship perceives an unknown transmission as a distress call...",
		Director: &Director{
			Firstname: "James",
			Lastname:  "Cameron",
		},
	})

	fs := http.FileServer(http.Dir("./public/"))
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))

	r.HandleFunc("/", getMovies).Methods("GET")
	r.HandleFunc("/movies/{id}", getMovie).Methods("GET")
	r.HandleFunc("/new-movie", newMovie).Methods("GET")
	r.HandleFunc("/edit-movie/{id}", editMovie).Methods("GET")

	r.HandleFunc("/create-movie", createMovie).Methods("POST")
	r.HandleFunc("/update-movie/{id}", updateMovie).Methods("PUT")
	r.HandleFunc("/delete-movie/{id}", deleteMovie).Methods("DELETE")

	fmt.Printf("Starting server on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", r))
}
