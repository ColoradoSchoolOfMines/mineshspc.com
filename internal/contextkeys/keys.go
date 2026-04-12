package contextkeys

type contextKey int

const (
	ContextKeyLoggedInTeacher contextKey = iota
	ContextKeyPageName
	ContextKeyRegistrationEnabled
	ContextKeyHostedByHTML
)
