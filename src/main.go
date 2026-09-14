package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

var templ *template.Template

type Exercise struct {
	Lift   string
	Weight int
	Reps   int
	Sets   int
}

var Exercises []Exercise

func main() {
	templ = template.Must(template.ParseGlob("templates/*.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("/workout-app", landingHandler)
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
		var exercise Exercise
		exercise.Lift = r.PostFormValue("lift")
		exercise.Weight, _ = strconv.Atoi(r.PostFormValue("weight"))
		exercise.Reps, _ = strconv.Atoi(r.PostFormValue("reps"))
		exercise.Sets, _ = strconv.Atoi(r.PostFormValue("sets"))
		Exercises = append(Exercises, exercise)
		http.Redirect(w, r, "/workout-new", http.StatusSeeOther)
		fmt.Println(Exercises)
	}
}

func activeHandler(w http.ResponseWriter, r *http.Request) {
	err := templ.ExecuteTemplate(w, "workout-active.html", nil)
	if err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	err := templ.ExecuteTemplate(w, "workout-log.html", nil)
	if err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}
