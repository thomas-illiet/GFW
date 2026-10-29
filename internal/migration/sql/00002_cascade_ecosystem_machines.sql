-- +goose Up
ALTER TABLE machines
    DROP CONSTRAINT machines_ecosystem_id_fkey;

ALTER TABLE machines
    ADD CONSTRAINT machines_ecosystem_id_fkey
    FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE machines
    DROP CONSTRAINT machines_ecosystem_id_fkey;

ALTER TABLE machines
    ADD CONSTRAINT machines_ecosystem_id_fkey
    FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id);
