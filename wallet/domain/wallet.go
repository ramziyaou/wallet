package domain

type Wallet struct {
	ID        int    `json:"id"`
	Ts        string `json:"ts"`
	UpdatedAt string `json:"updatedAt"`
	AccountNo string `json:"accountno"`
	IIN       string `json:"iin"`
	Transfer  int    `json:"transfer"`
	Amount    int    `json:"amount"`
}
