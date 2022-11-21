# Mbti Test

## Build & run

`API_KEY` is a valid API key for the tmdb API: `API_KEY=xxx go run .`

A terminal with support for ANSI codes is required. That should be any terminal emulator of this century. It has been tested to work with the default terminal emulator on macOS, and with others (e.g. Alacritty).


## Code notes

The API requires a few separate calls. Each API endpoint's response is deserialized into a separate struct. Most fields are not useful to us and skipped.

The only data structure we use is an array of structs, containing a few hundred elements, which we linearly scan in a few instances, because it's simple, fast and does not put pressure on the garbage collector.

Failure to contact the API will halt the application. A retry mechanism is easy to add, e.g. https://pkg.go.dev/github.com/cenkalti/backoff/v4#example-Retry .

The UI is a simple textual interface in the terminal, using the basic library `pterm` to implement that quickly. It's just ANSI codes. It should work with different color themes (dark/light mode), since it has been tested with both.

There are no tests. That's because of the time constraint and also because there is barely any logic and the application stops on most errors so there is no complicated structure to the code.

One testing approach would be to mock the API by passing a function pointer to the functions instead of them calling the API over HTTP directly. That can be useful to trigger timeouts, JSON parsing errors, etc.

About the architecture: it's just a few functions operating on structs. It's very C like. That's what I found to be the simplest to understand in my software engineer career, and is very easy to debug (no virtual calls/indirections). People coming from any programming language will comprehend it right away.
