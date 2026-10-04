-- +goose Up
CREATE TABLE workouts (
    id INTEGER PRIMARY KEY,
    finished_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE workout_exercises (
    id INTEGER PRIMARY KEY,
    workout_id INTEGER NOT NULL
        REFERENCES workouts(id) ON DELETE CASCADE,
    lift_name TEXT NOT NULL,
    position INTEGER NOT NULL CHECK (position >= 1),
    UNIQUE (workout_id, position)
);

CREATE TABLE workout_sets (
    id INTEGER PRIMARY KEY,
    workout_exercise_id INTEGER NOT NULL
        REFERENCES workout_exercises(id) ON DELETE CASCADE,
    set_number INTEGER NOT NULL CHECK (set_number >= 1),
    weight INTEGER NOT NULL CHECK (weight >= 0),
    reps INTEGER NOT NULL CHECK (reps >= 0),
    completed INTEGER NOT NULL DEFAULT 0
        CHECK (completed IN (0, 1)),
    UNIQUE (workout_exercise_id, set_number)
);

-- +goose Down
DROP TABLE workout_sets;
DROP TABLE workout_exercises;
DROP TABLE workouts;
