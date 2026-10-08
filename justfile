default:
	go build -C ./src -o ../bin/out

run:
	go build -C ./src -o ../bin/out
	bin/out

migrate_up:
	goose -dir sql/schema sqlite3 ./workouts.db up

migrate_down:
	goose -dir sql/schema sqlite3 ./workouts.db down

clean_db:
    sqlite3 workouts.db "delete from workouts;"
    sqlite3 workouts.db "delete from workout_exercises;"
    sqlite3 workouts.db "delete from workout_sets;"
