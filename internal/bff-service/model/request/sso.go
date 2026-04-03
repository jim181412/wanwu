package request

type SSOLogin struct {
	CallbackURL string `form:"callbackUrl" json:"callbackUrl"`
	MockUser    string `form:"mockUser" json:"mockUser"`
}

func (r *SSOLogin) Check() error {
	return nil
}

type SSOExchange struct {
	CallbackURL string `form:"callbackUrl" json:"callbackUrl"`
	Ticket      string `form:"ticket" json:"ticket"`
}

func (r *SSOExchange) Check() error {
	return nil
}

type SSOLogout struct {
	CallbackURL string `form:"callbackUrl" json:"callbackUrl"`
}

func (r *SSOLogout) Check() error {
	return nil
}
