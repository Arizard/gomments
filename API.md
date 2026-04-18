# API Endpoints

All endpoints are prefixed with `BASE_URL` if configured.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/ping` | Health check, returns "pong" |
| GET | `/articles/:article/replies` | Get all comments for an article |
| POST | `/articles/:article/replies` | Submit a new comment to an article |
| GET | `/articles/replies/stats` | Get comment counts for multiple articles (use `?article=` query params) |
| POST | `/articles/:article/reactions/like` | Add a "like" reaction to an article |
| DELETE | `/reactions` | Delete a reaction by the deletion key (use `?key=` query param) |
| GET | `/articles/reactions/stats` | Get reaction counts for multiple articles (use `?article=` query params) |
