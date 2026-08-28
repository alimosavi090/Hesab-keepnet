package enums

type Currency string

const (
	CurrencyRIAL Currency = "RIAL"
	CurrencyUSD  Currency = "USD"
)

func (c Currency) Valid() bool {
	return c == CurrencyRIAL || c == CurrencyUSD
}

type CategoryType string

const (
	CategoryBusiness CategoryType = "BUSINESS"
	CategoryPersonal CategoryType = "PERSONAL"
)

func (t CategoryType) Valid() bool {
	return t == CategoryBusiness || t == CategoryPersonal
}

type Gateway string

const (
	GatewayZarinpal   Gateway = "ZARINPAL"
	GatewayCardToCard Gateway = "CARD_TO_CARD"
	GatewaySupport    Gateway = "SUPPORT"
)

func (g Gateway) Valid() bool {
	return g == GatewayZarinpal || g == GatewayCardToCard || g == GatewaySupport
}

type LedgerType string

const (
	LedgerIncome      LedgerType = "INCOME"
	LedgerExpense     LedgerType = "EXPENSE"
	LedgerTransferIn  LedgerType = "TRANSFER_IN"
	LedgerTransferOut LedgerType = "TRANSFER_OUT"
)

func (t LedgerType) Valid() bool {
	switch t {
	case LedgerIncome, LedgerExpense, LedgerTransferIn, LedgerTransferOut:
		return true
	}
	return false
}

func (t LedgerType) IsOutflow() bool {
	return t == LedgerExpense || t == LedgerTransferOut
}

type RepDirection string

const (
	RepDebit  RepDirection = "DEBIT"
	RepCredit RepDirection = "CREDIT"
)

func (d RepDirection) Valid() bool {
	return d == RepDebit || d == RepCredit
}

type RepeatInterval string

const (
	RepeatNone    RepeatInterval = "NONE"
	RepeatMonthly RepeatInterval = "MONTHLY"
	RepeatYearly  RepeatInterval = "YEARLY"
)

func (r RepeatInterval) Valid() bool {
	return r == RepeatNone || r == RepeatMonthly || r == RepeatYearly
}

var DefaultCategories = map[CategoryType][]string{
	CategoryBusiness: {
		"سرور ایران",
		"سرور خارج",
		"ترافیک سرور ایران",
		"هاست و دامنه",
		"هزینه اضطراری",
	},
	CategoryPersonal: {
		"کافه",
		"غذا بیرون",
		"هزینه روزانه",
		"هزینه سگ",
		"سمیرا",
		"وام",
		"کرایه خانه",
	},
}
