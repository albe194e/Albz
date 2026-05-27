package ui

type Page int

const (
	Landing  Page = 0
	Login    Page = 1
	Register Page = 2
	Chat     Page = 3
)

type UIState struct {
	Page Page
}
