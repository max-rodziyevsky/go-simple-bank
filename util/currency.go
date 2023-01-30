package util

const (
	USD = "USD"
	EUR = "EUR"
	UAH = "UAH"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case EUR, USD, UAH:
		return true
	}
	return false
}
