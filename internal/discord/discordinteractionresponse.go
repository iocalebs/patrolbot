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
	Type InteractionType `json:"type"`
}
