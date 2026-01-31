package discord

// InteractionType indicates the type of the Interaction.
// See: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-type
type InteractionType uint16

// Interaction type codes.
// See: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-type
const (
	InteractionTypePing = iota + 1
	InteractionTypeApplicationCommand
)

// Interaction represents an Interaction object.
// See: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object
type Interaction struct {
	Type InteractionType                   `json:"type"`
	Data ApplicationCommandInteractionData `json:"data"`
}

// ApplicationCommandInteractionData represents
// the data field in an [Interaction] when the interaction type is APPLICATION_COMMAND.
// See: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type ApplicationCommandInteractionData struct {
	Name    string                                `json:"name"`
	GuildID string                                `json:"guild_id"`
	Options []ApplicationCommandInteractionOption `json:"options"`
}

// ApplicationCommandInteractionOption represents command arguments from the user.
// See: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-interaction-data-option-structure
type ApplicationCommandInteractionOption struct {
	Name string `json:"name"`
}
