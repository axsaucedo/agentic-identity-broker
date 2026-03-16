package id

// ClientID is an OAuth2 client identifier (VARCHAR(255), not a UUID).
type ClientID string

func NewClientID(s string) ClientID { return ClientID(s) }
func (id ClientID) String() string  { return string(id) }
func (id ClientID) IsZero() bool    { return id == "" }

// ExternalID is an optional external governance system identifier (VARCHAR(255)).
type ExternalID string

func NewExternalID(s string) ExternalID { return ExternalID(s) }
func (id ExternalID) String() string    { return string(id) }
func (id ExternalID) IsZero() bool      { return id == "" }

// Principal is an authenticated user identity string (email, subject).
type Principal string

func NewPrincipal(s string) Principal { return Principal(s) }
func (id Principal) String() string   { return string(id) }
func (id Principal) IsZero() bool     { return id == "" }
