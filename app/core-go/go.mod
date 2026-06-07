module github.com/albe194e/albz/app/core-go

go 1.26.3

require (
	github.com/albe194e/albz/shared v0.0.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	modernc.org/sqlite v1.50.1
)

replace github.com/albe194e/albz/shared => ../../shared
