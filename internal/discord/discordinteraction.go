package discord

// InteractionType indicates the type of the Interaction
// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-type
type InteractionType uint16

// Interaction type codes
// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-type
const (
	InteractionTypePing = iota + 1
	InteractionTypeApplicationCommand
)

// Interaction represents an Interaction object
// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object
type Interaction struct {
	Type InteractionType `json:"type"`
}
