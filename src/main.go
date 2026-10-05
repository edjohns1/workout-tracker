package main

import (
	"database/sql"
	"fmt"
	"html/template"
	_ "modernc.org/sqlite"
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

	// open db and test connection
	db, err := sql.Open("sqlite", "./workouts.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(err)
	}

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
			return
		}
	} else if r.Method == "POST" {
		// pull values from form
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Could not render page", http.StatusInternalServerError)
			return
		}
		var finishedExercises []Exercise
		for i, exercise := range Exercises {
			var finishedSets []Set
			for j := range exercise.Sets {
				weightText := fmt.Sprintf("weight_%d_%d", i, j)
				repsText := fmt.Sprintf("reps_%d_%d", i, j)
				completedText := fmt.Sprintf("completed_%d_%d", i, j)

				weightVal := r.PostFormValue(weightText)
				repsVal := r.PostFormValue(repsText)
				_, completed := r.PostForm[completedText]

				weight, err := strconv.Atoi(weightVal)
				if err != nil {
					http.Error(w, "Could not render page", http.StatusBadRequest)
					return
				}
				reps, err := strconv.Atoi(repsVal)
				if err != nil {
					http.Error(w, "Could not render page", http.StatusBadRequest)
					return
				}
				s := Set{
					Weight:    weight,
					Reps:      reps,
					Completed: completed,
				}
				finishedSets = append(finishedSets, s)
			}
			e := Exercise{
				Lift: exercise.Lift,
				Sets: finishedSets,
			}
			finishedExercises = append(finishedExercises, e)
		}
		// submit to DB
		fmt.Println(finishedExercises)

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
