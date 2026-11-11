package gopdf

type InvoiceItem struct {
	Name      string  `json:"name"`
	Qty       int     `json:"qty"`
	UnitPrice float64 `json:"unit_price"`
	Amount    float64 `json:"amount"`
}

type InvoiceData struct {
	InvoiceNumber string        `json:"invoice_number"`
	Date          string        `json:"date"`
	CompanyName   string        `json:"company_name"`
	CompanyAddress string       `json:"company_address"`
	CustomerName  string        `json:"customer_name"`
	CustomerPhone string        `json:"customer_phone"`
	Items         []InvoiceItem `json:"items"`
	Subtotal      float64       `json:"subtotal"`
	TaxAmount     float64       `json:"tax_amount"`
	TotalAmount   float64       `json:"total_amount"`
}

type ReceiptItem struct {
	Name   string  `json:"name"`
	Qty    int     `json:"qty"`
	Amount float64 `json:"amount"`
}

type ReceiptData struct {
	ShopName      string        `json:"shop_name"`
	ShopAddress   string        `json:"shop_address"`
	ShopPhone     string        `json:"shop_phone"`
	BillNumber    string        `json:"bill_number"`
	Date          string        `json:"date"`
	CustomerName  string        `json:"customer_name,omitempty"`
	Items         []ReceiptItem `json:"items"`
	Subtotal      float64       `json:"subtotal"`
	TaxAmount     float64       `json:"tax_amount"`
	TotalAmount   float64       `json:"total_amount"`
	PaymentMethod string        `json:"payment_method"`
}
