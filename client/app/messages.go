package app

import (
	"fmt"
)

func (c *Controller) AddMessage() {

	fmt.Println("Sending message")
	/*
		m := sql.Message{
			Body: msg,
		}

		err := c.Store.Q.CreateMessage(context.Background(), sql.CreateMessageParams{
			SenderID:       "user1",
			RecipientID:    "user2",
			Body:           m.Body,
			ConversationID: "conv1",
			CreatedAt:      0,
			Direction:      "outgoing",
			DeliveryState:  "pending",
		})
		if err != nil {
			// Handle error
		}
	*/
}
