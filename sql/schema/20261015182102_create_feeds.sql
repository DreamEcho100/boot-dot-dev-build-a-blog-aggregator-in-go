-- +goose Up
CREATE TABLE
	feeds (
		id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		url VARCHAR(255) UNIQUE NOT NULL,
		user_id uuid NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);

CREATE TRIGGER set_feeds_updated_at BEFORE
UPDATE ON feeds FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS set_feeds_updated_at ON feeds;

DROP TABLE IF EXISTS feeds;
