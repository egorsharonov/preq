package containers

type GetByType uint8

const (
	OrderID GetByType = iota
	CdbProcessID
)

func (e GetByType) String() string {
	switch e {
	case OrderID:
		return "Order ID"
	case CdbProcessID:
		return "CDB process ID"
	default:
		return "invalid GetBy type"
	}
}

type OrderGetByContainer struct {
	Key       int64
	GetByType GetByType
	ForUpdate bool
}
