package gopdfkit

// InvoiceItem represents a line item on an invoice.
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

// Invoice holds complete data for rendering A4 / A5 corporate and retail invoices.
type Invoice struct {
	InvoiceNumber   string        `json:"invoice_number"`
	Date            string        `json:"date"`
	DueDate         string        `json:"due_date,omitempty"`
	CompanyName     string        `json:"company_name"`
	CompanyAddress  string        `json:"company_address,omitempty"`
	CompanyPhone    string        `json:"company_phone,omitempty"`
	CompanyGST      string        `json:"company_gst,omitempty"`
	CustomerName    string        `json:"customer_name"`
	CustomerPhone   string        `json:"customer_phone,omitempty"`
	CustomerAddress string        `json:"customer_address,omitempty"`
	Items           []InvoiceItem `json:"items"`
	Subtotal        float64       `json:"subtotal"`
	Tax             float64       `json:"tax"`
	Discount        float64       `json:"discount,omitempty"`
	GrandTotal      float64       `json:"grand_total"`
	Currency        string        `json:"currency"` // e.g. "₹", "$", "INR"
	PaymentMode     string        `json:"payment_mode,omitempty"`
	Notes           string        `json:"notes,omitempty"`
}

// ReceiptItem represents a line item for POS thermal slips.
type ReceiptItem struct {
	Name   string  `json:"name"`
	Qty    int     `json:"qty"`
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}

// Receipt holds data for 80mm / 58mm thermal counter roll receipts.
type Receipt struct {
	StoreName    string        `json:"store_name"`
	StoreAddress string        `json:"store_address,omitempty"`
	StorePhone   string        `json:"store_phone,omitempty"`
	ReceiptNo    string        `json:"receipt_no"`
	Date         string        `json:"date"`
	Cashier      string        `json:"cashier,omitempty"`
	Items        []ReceiptItem `json:"items"`
	Subtotal     float64       `json:"subtotal"`
	Tax          float64       `json:"tax,omitempty"`
	GrandTotal   float64       `json:"grand_total"`
	Currency     string        `json:"currency"`
	PaymentMode  string        `json:"payment_mode,omitempty"`
	FooterNote   string        `json:"footer_note,omitempty"`
}
