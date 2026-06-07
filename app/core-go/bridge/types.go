package bridge

const EventTypeStateChanged = "state_changed"

type Options struct {
	ProfileName string
	DataDir     string
	ServerURL   string
}

type Config struct {
	ProfileName string `json:"profile_name"`
	DataDir     string `json:"data_dir"`
	DBPath      string `json:"db_path"`
	ServerURL   string `json:"server_url"`
}

type EventSink interface {
	OnCoreEvent(eventJSON string)
}

type Event struct {
	Type     string    `json:"type"`
	Snapshot *Snapshot `json:"snapshot,omitempty"`
}

type Snapshot struct {
	CurrentUser          *User           `json:"current_user,omitempty"`
	Messages             []Message       `json:"messages"`
	Conversations        []Conversation  `json:"conversations"`
	Friends              []Friend        `json:"friends"`
	FriendRequests       []FriendRequest `json:"friend_requests"`
	LoadedConversationID string          `json:"loaded_conversation_id"`
	ServerConnected      bool            `json:"server_connected"`
	LastNetworkError     string          `json:"last_network_error"`
}

type User struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Username          string `json:"username"`
	ProfilePictureUrl string `json:"profile_picture_url"`
	FriendCode        string `json:"friend_code"`
}

type Conversation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Message struct {
	ID              int64  `json:"id"`
	ConversationID  string `json:"conversation_id"`
	SenderID        string `json:"sender_id"`
	ClientMessageID string `json:"client_message_id"`
	Body            string `json:"body"`
	CreatedAt       int64  `json:"created_at"`
	DeliveryState   string `json:"delivery_state"`
}

type Friend struct {
	ID                int64  `json:"id"`
	UserID            string `json:"user_id"`
	Name              string `json:"name"`
	Username          string `json:"username"`
	ProfilePictureUrl string `json:"profile_picture_url"`
	FriendCode        string `json:"friend_code"`
	CreatedAt         int64  `json:"created_at"`
}

type FriendRequest struct {
	ID             int64  `json:"id"`
	FromUserID     string `json:"from_user_id"`
	Name           string `json:"name"`
	Username       string `json:"username"`
	FromFriendCode string `json:"from_friend_code"`
	CreatedAt      int64  `json:"created_at"`
}
