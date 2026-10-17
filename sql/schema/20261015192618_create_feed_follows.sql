-- +goose Up
CREATE TABLE feed_follows (
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  user_id uuid NOT NULL,
  feed_id uuid NOT NULL,

  PRIMARY KEY (feed_id, user_id),
  FOREIGN KEY (feed_id) REFERENCES feeds (id ) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users (id ) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feed_follows;
