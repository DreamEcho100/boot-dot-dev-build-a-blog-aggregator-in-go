-- name: CreatePost :one
INSERT INTO posts (title, description, url, feed_id, published_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, title, description, url, feed_id, published_at, created_at, updated_at;

-- name: GetPostsByUserId :many
SELECT
  posts.id,
  posts.title,
  posts.description,
  posts.url,
  posts.feed_id,
  posts.published_at,
  posts.created_at,
  posts.updated_at
FROM posts
JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
WHERE feed_follows.user_id = $1
ORDER BY posts.created_at DESC
LIMIT $2;
