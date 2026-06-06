package ui

import "fyne.io/fyne/v2/data/binding"

type Page int

const (
	Landing  Page = 0
	Login    Page = 1
	Register Page = 2
	Chat     Page = 3
	Profile  Page = 4
)

type UIState struct {
	Page                   Page
	ActiveConversationID   string
	ActiveConversationName binding.String
	SidebarOpen            bool
}

func (s *UIState) Init() {
	s.Page = Landing
	s.ActiveConversationID = ""
	s.ActiveConversationName = binding.NewString()
	_ = s.ActiveConversationName.Set("Choose a conversation")
	s.SidebarOpen = false
}
