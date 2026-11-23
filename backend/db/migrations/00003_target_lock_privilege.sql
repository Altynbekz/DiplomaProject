-- +goose Up
-- PostgreSQL SELECT FOR UPDATE requires UPDATE on at least one column.
-- Only the display name is mutable by the runtime role; origin/identity and evidence remain immutable.
GRANT UPDATE(name) ON targets TO diploma_runtime;
-- +goose Down
REVOKE UPDATE(name) ON targets FROM diploma_runtime;
