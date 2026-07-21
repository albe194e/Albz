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
	CurrentUser          *User            `json:"current_user,omitempty"`
	Messages             []Message        `json:"messages"`
	Conversations        []Conversation   `json:"conversations"`
	Contacts             []Contact        `json:"contacts"`
	ContactRequests      []ContactRequest `json:"contact_requests"`
	LoadedConversationID string           `json:"loaded_conversation_id"`
	ServerConnected      bool             `json:"server_connected"`
	LastNetworkError     string           `json:"last_network_error"`
}

type User struct {
	ID                 string `json:"id"`
	UserID             string `json:"user_id"`
	DeviceID           string `json:"device_id"`
	Name               string `json:"name"`
	LocalHandle        string `json:"local_handle"`
	Username           string `json:"username"`
	ProfilePicturePath string `json:"profile_picture_path"`
	ProfilePictureUrl  string `json:"profile_picture_url"`
	ContactCode        string `json:"contact_code"`
}

type Conversation struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type Message struct {
	ID              string `json:"id"`
	ConversationID  string `json:"conversation_id"`
	SenderUserID    string `json:"sender_user_id"`
	SenderID        string `json:"sender_id"`
	SenderDeviceID  string `json:"sender_device_id"`
	ClientMessageID string `json:"client_message_id"`
	Body            string `json:"body"`
	CreatedAt       int64  `json:"created_at"`
	ReceivedAt      int64  `json:"received_at"`
	Direction       string `json:"direction"`
	DeliveryState   string `json:"delivery_state"`
}

type Contact struct {
	ID                 int64  `json:"id"`
	UserID             string `json:"user_id"`
	DisplayName        string `json:"display_name"`
	Name               string `json:"name"`
	LocalHandle        string `json:"local_handle"`
	Username           string `json:"username"`
	ProfilePicturePath string `json:"profile_picture_path"`
	ProfilePictureUrl  string `json:"profile_picture_url"`
	ContactCode        string `json:"contact_code"`
	CreatedAt          int64  `json:"created_at"`
}

type ContactRequest struct {
	ID                 int64  `json:"id"`
	FromUserID         string `json:"from_user_id"`
	FromDeviceID       string `json:"from_device_id"`
	DisplayName        string `json:"display_name"`
	Name               string `json:"name"`
	LocalHandle        string `json:"local_handle"`
	Username           string `json:"username"`
	ProfilePicturePath string `json:"profile_picture_path"`
	FromContactCode    string `json:"from_contact_code"`
	InvitePayload      string `json:"invite_payload"`
	State              string `json:"state"`
	CreatedAt          int64  `json:"created_at"`
}
