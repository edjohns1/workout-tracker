package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
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

type WorkoutQueryResult struct {
	Finished_at string
	Workout_id  int
	Lift_id     int
	Lift_name   string
	Set_id      int
	Weight      int
	Reps        int
}

type WorkoutCard struct {
	ID          int
	Finished_at string
	Sets        []WorkoutQueryResult
}

type WorkoutResult struct {
	Finished_at string
	Exercise    []Exercise
}

var Exercises []Exercise
var db *sql.DB

func main() {
	var err error

	// open db and test connection
	db, err = sql.Open("sqlite", "./workouts.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(err)
	}

	templ = template.Must(template.ParseGlob("templates/*.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /app", landingHandler)
	mux.HandleFunc("/workout-new", newHandler)
	mux.HandleFunc("/workout-active", activeHandler)
	mux.HandleFunc("GET /workout-log", logHandler)

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
		err := templ.ExecuteTemplate(w, "workout-new.html", Exercises)
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

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			http.Error(w, "Could not submit workout", http.StatusInternalServerError)
			return
		}

		workout_insert_result, err := tx.ExecContext(r.Context(), "INSERT INTO workouts DEFAULT VALUES")
		if err != nil {
			http.Error(w, "Could not submit workout", http.StatusInternalServerError)
			return
		}
		workout_id, _ := workout_insert_result.LastInsertId()

		var finishedExercises []Exercise
		for i, exercise := range Exercises {

			exercise_insert_result, err := tx.ExecContext(r.Context(),
				`INSERT INTO workout_exercises
    		     (workout_id, lift_name, position) VALUES (?, ?, ?)`,
				workout_id, exercise.Lift, i+1)
			if err != nil {
				http.Error(w, "Could not render page", http.StatusBadRequest)
				return
			}
			exercise_id, _ := exercise_insert_result.LastInsertId()

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

				completedInt := 0
				if completed {
					completedInt = 1
				}

				_, err = tx.ExecContext(r.Context(),
					`INSERT INTO workout_sets
             		 (workout_exercise_id, set_number, weight, reps, completed) VALUES (?, ?, ?, ?, ?)`,
					exercise_id, j+1, s.Weight, s.Reps, completedInt)

				if err != nil {
					http.Error(w, "Could not render page", http.StatusBadRequest)
					return
				}
			}
			e := Exercise{
				Lift: exercise.Lift,
				Sets: finishedSets,
			}
			finishedExercises = append(finishedExercises, e)
		}

		tx.Commit()

		Exercises = nil
		http.Redirect(w, r, "/workout-log", http.StatusSeeOther)
	}
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	statement := `
WITH recent AS
(SELECT * FROM workouts ORDER BY finished_at DESC LIMIT 3)
SELECT recent.finished_at, recent.id, workout_exercises.id, workout_exercises.lift_name, workout_sets.id, workout_sets.weight, workout_sets.reps FROM recent
JOIN workout_exercises on recent.id=workout_exercises.workout_id
JOIN workout_sets on workout_exercises.id=workout_sets.workout_exercise_id
ORDER BY recent.finished_at DESC, recent.id DESC, workout_exercises.id, workout_sets.id;`
	rows, _ := db.Query(statement)
	defer rows.Close()

	var cards []WorkoutCard
	i := -1
	for rows.Next() {
		var wr WorkoutQueryResult
		err = rows.Scan(
			&wr.Finished_at,
			&wr.Workout_id,
			&wr.Lift_id,
			&wr.Lift_name,
			&wr.Set_id,
			&wr.Weight,
			&wr.Reps,
		)
		if err != nil {
			fmt.Println(err)
		}
		if len(cards) == 0 || cards[len(cards)-1].ID != wr.Workout_id {
			cards = append(cards, WorkoutCard{
				ID:          wr.Workout_id,
				Finished_at: wr.Finished_at,
			})
			i += 1
		}
		cards[i].Sets = append(cards[i].Sets, wr)
	}
	// fmt.Printf("finished: %s, lift: %s\nweight: %d, reps:%d\n", wr.finished_at, wr.lift_name, wr.weight, wr.reps)
	// fmt.Println("-------\n")
	err = templ.ExecuteTemplate(w, "workout-log.html", cards)
	if err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}
