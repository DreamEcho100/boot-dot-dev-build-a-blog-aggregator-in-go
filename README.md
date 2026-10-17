# Gator

Gator is a Go CLI that fetches RSS feeds and stores their posts in PostgreSQL. You can create users, add and follow feeds, run the aggregator, and browse saved posts.

## Requirements

- Go 1.23.5 or later
- PostgreSQL
- [Goose](https://github.com/pressly/goose) to apply database migrations

Install Goose:

```
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Make sure Go's binary directory is on your `PATH` so you can run installed tools.

## Database setup

Create a PostgreSQL database named `gator`. For example:

```
createdb -h localhost -U postgres gator
```

From the root of this repository, apply the migrations:

```
goose -dir ./sql/schema postgres "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable" up
```

Adjust the username, password, host, and database name for your PostgreSQL setup. The database URL in the next step **must point to the same database** you migrated.

## Configuration

Create `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Gator updates `current_user_name` when you register or log in.

## Install and run

Install the CLI with Go:

```
go install github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go@latest
```

With the current module name, the installed executable is named `boot-dot-dev-build-a-blog-aggregator-in-go`. For example:

```
boot-dot-dev-build-a-blog-aggregator-in-go register alice
```

If you want an executable named **`gator`**, clone the repository and build it from the project root:

```
go build -o gator .
```

Then run it:

```
./gator register alice
```

You need Go to build or install Gator, but **not** to run the resulting binary. `go run .` is useful during development; the installed or built binary is what you run afterward.

## Commands

Examples below use `gator` as the executable name. If you used `go install`, substitute the longer installed executable name.

| Command                                                          | Description                                                                      |
| ---------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `gator register alice`                                           | Create a user and set them as the current user.                                  |
| `gator login alice`                                              | Switch to an existing user.                                                      |
| `gator users`                                                    | List users.                                                                      |
| `gator addfeed "Hacker News" "https://news.ycombinator.com/rss"` | Add a feed and automatically follow it.                                          |
| `gator feeds`                                                    | List available feeds.                                                            |
| `gator follow "https://news.ycombinator.com/rss"`                | Follow an existing feed.                                                         |
| `gator following`                                                | List feeds followed by the current user.                                         |
| `gator unfollow "https://news.ycombinator.com/rss"`              | Stop following a feed.                                                           |
| `gator agg 1m`                                                   | Continuously fetch feeds, waiting one minute between requests. Stop with Ctrl+C. |
| `gator browse`                                                   | Browse saved posts, if supported by your current version.                        |
| `gator reset`                                                    | Delete users and their associated feeds, follows, and posts. **Destructive.**    |

Run `agg` in one terminal, then use commands such as `browse` in another. Use a reasonable aggregation interval so you don't send excessive requests to feed servers.

## Development

After changing queries in `sql/queries`, regenerate the database code with:

```
sqlc generate
```

Run Go tests from the repository root:

```
go test ./...
```

The aggregation integration test calls `reset` against the database supplied in `GATOR_TEST_DATABASE_URL`. **Do not point it at a database containing data you want to keep.** Run only that test with:

```
GATOR_TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/gator?sslmode=disable" go test . -run '^TestAggAndBrowseIntegration$' -count=1 -v
```

## Extending the Project

- Add sorting and filtering options to the browse command
- Add pagination to the browse command
- Add concurrency to the agg command so that it can fetch more frequently
- Add a search command that allows for fuzzy searching of posts
- Add bookmarking or liking posts
- Add a TUI that allows you to select a post in the terminal and view it in a more readable format (either in the terminal or open in a browser)
- Add an HTTP API (and authentication/authorization) that allows other users to interact with the service remotely
- Write a service manager that keeps the agg command running in the background and restarts it if it crashes
