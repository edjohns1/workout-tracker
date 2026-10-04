package main

import (
	"html/template"
	"net/http"
	"strconv"
)

var templ *template.Template

type Set struct {
	Weight    int
	Reps      int
	Completed bool
}

type Exercise struct {
	Lift string
	Sets []Set
}

var Exercises []Exercise

func main() {
	templ = template.Must(template.ParseGlob("templates/*.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("/app", landingHandler)
	mux.HandleFunc("/workout-new", newHandler)
	mux.HandleFunc("/workout-active", activeHandler)
	mux.HandleFunc("/workout-log", logHandler)

	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/style.css")
	})

	http.ListenAndServe(":8080", mux)
}

func landingHandler(w http.ResponseWriter, r *http.Request) {
	err := templ.ExecuteTemplate(w, "workout-app.html", nil)
	if err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := templ.ExecuteTemplate(w, "workout-new.html", nil)
		if err != nil {
			http.Error(w, "Could not render page", http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		lift := r.PostFormValue("lift")
		weight, _ := strconv.Atoi(r.PostFormValue("weight"))
		reps, _ := strconv.Atoi(r.PostFormValue("reps"))
		setCount, _ := strconv.Atoi(r.PostFormValue("sets"))

		exercise := Exercise{Lift: lift}
		for i := 0; i < setCount; i++ {
			exercise.Sets = append(exercise.Sets, Set{
				Weight:    weight,
				Reps:      reps,
				Completed: false,
			})
		}
		Exercises = append(Exercises, exercise)
		http.Redirect(w, r, "/workout-new", http.StatusSeeOther)
	}
}

func activeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := templ.ExecuteTemplate(w, "workout-active.html", Exercises)
		if err != nil {
			http.Error(w, "Could not render page", http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		// TODO add write to db
		Exercises = nil
		http.Redirect(w, r, "/workout-log", http.StatusSeeOther)
	}
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	err := templ.ExecuteTemplate(w, "workout-log.html", nil)
	if err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}
