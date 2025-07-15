# Gomments
A commenting system for my personal website less.coffee.

It's dead simple and doesn't require registration, or even email. Commenters can identify themselves with a signature (a bit like a password) which is converted to a unique hash displayed next to their name.

## Features

* Submit a `reply` which is associated with an `article` using `POST /articles/<article>/replies`
* Name and Tripcode is optional, default is 'Anonymous'
* List replies for an `article` in reverse chronological order using `GET /articles/<article>/replies`
* List stats for an `article` using `GET /articles/replies/stats?article=<article1>&article=<article2>`
* No external database needed, uses SQLite.
* Dockerfile included (but no official container registry image).
* Configure base path e.g environment can have `BASE_URL=/gomments` so that all routes are prefixed with `/gomments`

## Screenshots

Submit a name (optional tripcode, delimited by `#`) and a message body. Under the hood, provide an article ID to associate the reply to the article. These can be retrieved as a list.

The rendering is up to you. Here I've customised the interface for less.coffee:

<img width="664" height="403" alt="Screenshot 2025-07-15 at 7 29 27 pm" src="https://github.com/user-attachments/assets/8afb6365-b1a5-447d-aaf8-853ac7cfd7ba" />
<img width="662" height="295" alt="Screenshot 2025-07-15 at 7 28 45 pm" src="https://github.com/user-attachments/assets/a53bbc93-0e74-45e8-b656-e44528c247d3" />
