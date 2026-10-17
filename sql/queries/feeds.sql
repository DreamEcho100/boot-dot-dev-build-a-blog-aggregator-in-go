-- name: CreateFeed :one
INSERT INTO feeds (id, name, url, user_id)
VALUES ($1, $2, $3, $4)
RETURNING id, name, url, user_id, created_at, updated_at;

-- name: GetFeeds :many
SELECT
  feeds.id,
  feeds.name,
  feeds.url,
  feeds.created_at,
  feeds.updated_at,
  users.name AS user_name
FROM feeds
JOIN users ON users.id = feeds.user_id;

-- name: CreateFeedFollowByFeedUrl :one
WITH inserted_feed_follow AS (
  INSERT INTO feed_follows (feed_id, user_id)
  -- VALUES ((SELECT feeds.id FROM feeds WHERE feeds.url = $2), $1)
  -- The following returns a tuble that will be inserted as values
  SELECT feeds.id, $1 FROM feeds WHERE feeds.url = $2
  RETURNING feed_id, user_id, created_at
)
SELECT
    inserted_feed_follow.feed_id,
    inserted_feed_follow.user_id,
    inserted_feed_follow.created_at as created_at,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN users ON users.id = inserted_feed_follow.user_id
LIMIT 1;

-- name: DeleteFeedFollowByUrl :exec
DELETE FROM feed_follows
WHERE
  feed_follows.user_id = $2 AND
  feed_follows.feed_id IN (SELECT feeds.id FROM feeds WHERE feeds.url = $1);

-- name: GetFeedFollowsForUser :many
SELECT
  feeds.id,
  feeds.name,
  feeds.url,
  feeds.created_at,
  feeds.updated_at,
  feed_author.name AS author_name,
  users.id AS follower_id
FROM users
JOIN feed_follows ON feed_follows.user_id = users.id
JOIN feeds ON feeds.id = feed_follows.feed_id
JOIN users feed_author ON feed_author.id = feeds.user_id
WHERE users.id = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = now()::TIMESTAMPTZ, updated_at = now()::TIMESTAMPTZ
WHERE feeds.id = $1;

-- name: GetNextFeedToFetch :one
SELECT
  feeds.id,
  feeds.name,
  feeds.url,
  feeds.user_id,
  feeds.created_at,
  feeds.updated_at,
  feeds.last_fetched_at
FROM feeds
ORDER BY feeds.last_fetched_at ASC NULLS FIRST, feeds.created_at ASC, feeds.id ASC
LIMIT 1;

-- name: ResetFeeds :exec
DELETE FROM feeds;
