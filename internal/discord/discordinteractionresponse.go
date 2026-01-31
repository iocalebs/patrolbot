package discord

// InteractionCallbackType indicates the type of the [InteractionResponse]
// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-interaction-callback-type
type InteractionCallbackType uint16

// Interaction Callback type codes
// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-interaction-callback-type
const (
	InteractionCallbackTypePong                     = 1
	InteractionCallbackTypeChannelMessageWithSource = 4
)

// InteractionResponse represents an Interaction Response object
// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object
type InteractionResponse struct {
	Type InteractionType          `json:"type"`
	Data *InteractionResponseData `json:"data,omitempty"`
}

// InteractionResponseData represents the callback data of an Interaction Response object
// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-interaction-callback-data-structure
type InteractionResponseData struct {
	Content string `json:"content"` // Message content
}
