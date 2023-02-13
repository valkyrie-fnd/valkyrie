package evolution

type RequestBase struct {
	SID    string `json:"sid" validate:"required"`
	UserID string `json:"userId" validate:"required"`
	UUID   string `json:"uuid,omitempty"`
}
type StandardResponse struct {
	Status         string `json:"status" validate:"required"`
	Balance        Amount `json:"balance" validate:"required"`
	Bonus          Amount `json:"bonus"`
	UUID           string `json:"uuid,omitempty"`
	Retransmission bool   `json:"retransmission,omitempty"`
}

type CheckRequest struct {
	RequestBase
	Channel struct {
		Type string `json:"type"`
	} `json:"channel"`
}

type CheckResponse struct {
	Status string `json:"status"`
	SID    string `json:"sid"`
	UUID   string `json:"uuid"`
}

type BalanceRequest struct {
	RequestBase
	Game     interface{} `json:"game"`
	Currency string      `json:"currency" validate:"required,len=3"`
}

type Game struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Details GameDetails
}

type GameDetails struct {
	Table GameTable `json:"table"`
}

type GameTable struct {
	ID  string `json:"id"`
	Vid string `json:"vid"`
}

type DebitRequest struct {
	RequestBase
	Currency    string      `json:"currency" validate:"required,len=3"`
	Game        Game        `json:"game" validate:"required"`
	Transaction Transaction `json:"transaction" validate:"required"`
}

type CreditRequest struct {
	RequestBase
	Currency    string      `json:"currency" validate:"required,len=3"`
	Game        Game        `json:"game" validate:"required"`
	Transaction Transaction `json:"transaction" validate:"required"`
}

type CancelRequest struct {
	RequestBase
	Transaction Transaction `json:"transaction"`
	Currency    string      `json:"currency" validate:"required,len=3"`
	Game        Game        `json:"game" validate:"required"`
}

type PromoTransaction struct {
	Type            string `json:"type"`
	ID              string `json:"id"`
	Amount          Amount `json:"amount"`
	VoucherID       string `json:"voucherId"`
	RemainingRounds int    `json:"remainingRounds"`
}

type PromoPayoutRequest struct {
	RequestBase
	Currency         string           `json:"currency" validate:"required,len=3"`
	Game             Game             `json:"game" validate:"required"`
	PromoTransaction PromoTransaction `json:"promoTransaction"`
}

type Transaction struct {
	ID     string `json:"id"`
	RefID  string `json:"refId"`
	Amount Amount `json:"amount"`
}

type Player struct {
	Group     *Group  `json:"group,omitempty"`
	Session   Session `json:"session" validate:"required"`
	ID        string  `json:"id" validate:"required"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Nickname  string  `json:"nickname,omitempty"`
	Country   string  `json:"country" validate:"required,len=2"`
	Language  string  `json:"language" validate:"required,len=2"`
	Currency  string  `json:"currency" validate:"required,len=3"`
	Update    bool    `json:"update" validate:"required"`
}

type Session struct {
	ID string `json:"id" validate:"required"`
	IP string `json:"ip" validate:"required"`
}

type Group struct {
	ID     string `json:"id"`
	Action string `json:"action" validate:"required"`
}

type ConfigBrand struct {
	ID   string `json:"id"`
	Skin string `json:"skin"`
}

type ConfigGame struct {
	Category  string      `json:"category,omitempty"`
	Interface string      `json:"interface,omitempty"`
	PlayMode  string      `json:"playMode,omitempty"`
	Table     ConfigTable `json:"table"`
}

type ConfigTable struct {
	ID   string `json:"id" validate:"required"`
	Seat int    `json:"seat,omitempty"`
}

type ConfigChannel struct {
	Wrapped bool `json:"wrapped" validate:"required"`
	Mobile  bool `json:"mobile,omitempty"`
}

type ConfigUrls struct {
	Cashier            string `json:"cashier,omitempty"`
	ResponsibleGaming  string `json:"responsibleGaming,omitempty"`
	Lobby              string `json:"lobby,omitempty"`
	SessionTimeout     string `json:"sessionTimeout,omitempty"`
	GameHistory        string `json:"gameHistory,omitempty"`
	RealityCheckURL    string `json:"realityCheckURL,omitempty"`
	RngGoLiveURL       string `json:"rngGoLiveURL,omitempty"`
	RngGoLiveURLMobile string `json:"rngGoLiveURLMobile,omitempty"`
	RngLobbyButton     string `json:"rngLobbyButton,omitempty"`
	RngCloseButton     string `json:"rngCloseButton,omitempty"`
	RngHomeButton      string `json:"rngHomeButton,omitempty"`
	RngSessionTimeout  string `json:"rngSessionTimeout,omitempty"`
	RngErrorHandling   string `json:"rngErrorHandling,omitempty"`
	SweSelfTest        string `json:"sweSelfTest,omitempty"`
	SweGameLimits      string `json:"sweSelfGameLimits,omitempty"`
	SweSelfExclusion   string `json:"sweSelfExclusion,omitempty"`
}

type Config struct {
	Urls    ConfigUrls    `json:"urls"`
	Brand   ConfigBrand   `json:"brand"`
	Game    ConfigGame    `json:"game,omitempty"`
	Channel ConfigChannel `json:"channel" validate:"required"`
}

// UserAuthenticationRequest Evo user authentication
type UserAuthenticationRequest struct {
	UUID   string `json:"uuid" validate:"required"`
	Player Player `json:"player" validate:"required"`
	Config Config `json:"config" validate:"required"`
}

type UserAuthenticationResponse struct {
	Entry         string `json:"entry,omitempty"`
	EntryEmbedded string `json:"entryEmbedded,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type UserAuthenticationErrorResponse struct {
	Errors []Error `json:"errors"`
}

// Generic error codes
const (
	G0  = "G.0"  // Could not authenticate, please review sent data and try again. If problem persists, contact customer support 	System error, should be retried, in case of constant occurrences should be reported to Evolution.
	G1  = "G.1"  // Unknown casino $casinoKey will be provided by Evolution.
	G2  = "G.2"  // Provided $apiToken for casino $casinoKey is incorrect $apiToken will be provided by Evolution.
	G3  = "G.3"  // Player session creation is not configured for casino $casinoKey $apiToken have not been configured on Evolution side.
	G4  = "G.4"  // Unable to issue token System error, should be retried, in case of constant occurrences should be reported to Evolution.
	G5  = "G.5"  // Unable to authenticate user System error, should be retried, in case of constant occurrences should be reported to Evolution.
	G6  = "G.6"  // Unable to create user System error, should be retried, in case of constant occurrences should be reported to Evolution.
	G7  = "G.7"  // Unable to save player data System error, should be retried, in case of constant occurrences should be reported to Evolution.
	G8  = "G.8"  // Unable to authenticate user due to: $status Most likely client system returned invalid $status.
	G9  = "G.9"  // Clients IP address have been rejected Provided to evolution client IP address for white listing is incorrect.
	G10 = "G.10" // Only httpclient:// or https:// or native:// or app:// URL schemes are supported if 'config.channel.wrapped' is false Custom URL schemes are only supported for wrapped requests.
	G11 = "G.11" // Misconfiguration Can't launch the game with table id $tableId due to a misconfiguration. Please, contact customer support. Reference No - $refNumber
)
