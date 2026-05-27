package resources

type InvoiceResponse struct {
	InvoiceID     string  `json:"invoice_id"`
	PaymentID     string  `json:"payment_id"`
	InvoiceNumber string  `json:"invoice_number"`
	IssuedAt      string  `json:"issued_at"`
	TotalAmount   float64 `json:"total_amount"`
	PDFURL        string  `json:"pdf_url"`
}
