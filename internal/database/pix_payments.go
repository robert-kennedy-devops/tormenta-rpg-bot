package database

import "database/sql"

// PIX PAYMENT OPERATIONS

type PixPaymentRecord struct {
	ID          int
	CharacterID int
	PackageID   string
	Diamonds    int
	AmountBRL   float64
	TxID        string
	Status      string
	MPPaymentID int64
	QRCode      string
	QRCodeB64   string
}

func CreatePixPayment(charID int, packageID string, diamonds int, amountBRL float64, txid string, mpPaymentID int64, qrCode string, qrCodeB64 string) error {
	// Pass NULL when mpPaymentID is 0 to avoid unique constraint violation
	// (AbacatePay doesn't use numeric payment IDs)
	var mpID interface{}
	if mpPaymentID != 0 {
		mpID = mpPaymentID
	}
	_, err := DB.Exec(`
		INSERT INTO pix_payments
			(character_id, package_id, diamonds, amount_brl, txid, status, mp_payment_id, qr_code, qr_code_b64, expires_at)
		VALUES ($1,$2,$3,$4,$5,'pending',$6,$7,$8, NOW() + INTERVAL '30 minutes')
	`, charID, packageID, diamonds, amountBRL, txid, mpID, qrCode, qrCodeB64)
	return err
}

// GetPixPaymentByMPID fetches a payment by Mercado Pago payment ID.
func GetPixPaymentByMPID(mpPaymentID int64) (*PixPaymentRecord, error) {
	p := &PixPaymentRecord{}
	err := DB.QueryRow(`
		SELECT id, character_id, package_id, diamonds, amount_brl, txid, status,
		       COALESCE(mp_payment_id, 0), COALESCE(qr_code, ''), COALESCE(qr_code_b64, '')
		FROM pix_payments WHERE mp_payment_id=$1
	`, mpPaymentID).Scan(&p.ID, &p.CharacterID, &p.PackageID, &p.Diamonds, &p.AmountBRL, &p.TxID, &p.Status, &p.MPPaymentID, &p.QRCode, &p.QRCodeB64)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func GetPixPayment(txid string) (*PixPaymentRecord, error) {
	p := &PixPaymentRecord{}
	err := DB.QueryRow(`
		SELECT id, character_id, package_id, diamonds, amount_brl, txid, status,
		       COALESCE(mp_payment_id, 0), COALESCE(qr_code, ''), COALESCE(qr_code_b64, '')
		FROM pix_payments WHERE txid=$1
	`, txid).Scan(&p.ID, &p.CharacterID, &p.PackageID, &p.Diamonds, &p.AmountBRL, &p.TxID, &p.Status,
		&p.MPPaymentID, &p.QRCode, &p.QRCodeB64)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func GetPendingPixPayments(charID int) ([]PixPaymentRecord, error) {
	rows, err := DB.Query(`
		SELECT id, character_id, package_id, diamonds, amount_brl, txid, status,
		       COALESCE(mp_payment_id, 0), COALESCE(qr_code, ''), COALESCE(qr_code_b64, '')
		FROM pix_payments
		WHERE character_id=$1 AND status='pending' AND expires_at > NOW()
		ORDER BY created_at DESC
	`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payments []PixPaymentRecord
	for rows.Next() {
		var p PixPaymentRecord
		if err := rows.Scan(&p.ID, &p.CharacterID, &p.PackageID, &p.Diamonds, &p.AmountBRL, &p.TxID, &p.Status,
			&p.MPPaymentID, &p.QRCode, &p.QRCodeB64); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payments, nil
}

// GetAllPendingPixPayments returns ALL pending payments (for background polling goroutine).
func GetAllPendingPixPayments() ([]PixPaymentRecord, error) {
	rows, err := DB.Query(`
		SELECT id, character_id, package_id, diamonds, amount_brl, txid, status,
		       COALESCE(mp_payment_id, 0), COALESCE(qr_code, ''), COALESCE(qr_code_b64, '')
		FROM pix_payments
		WHERE status='pending' AND expires_at > NOW() AND mp_payment_id IS NOT NULL
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payments []PixPaymentRecord
	for rows.Next() {
		var p PixPaymentRecord
		if err := rows.Scan(&p.ID, &p.CharacterID, &p.PackageID, &p.Diamonds, &p.AmountBRL, &p.TxID, &p.Status,
			&p.MPPaymentID, &p.QRCode, &p.QRCodeB64); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payments, nil
}

// ConfirmPixPaymentByMPID marks a payment as paid using the Mercado Pago payment ID.
// Called by the MP webhook and the background polling goroutine.
func ConfirmPixPaymentByMPID(mpPaymentID int64) (charID, diamonds int, err error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Atomic idempotency: only one concurrent confirmer can flip pending -> paid.
	var p PixPaymentRecord
	err = tx.QueryRow(`
		UPDATE pix_payments
		SET status='paid', paid_at=NOW()
		WHERE mp_payment_id=$1 AND status='pending'
		RETURNING id, character_id, diamonds
	`, mpPaymentID).Scan(&p.ID, &p.CharacterID, &p.Diamonds)
	if err == sql.ErrNoRows {
		// Already processed (or missing payment). Keep compatibility by returning charID when found.
		err = tx.QueryRow(`SELECT character_id FROM pix_payments WHERE mp_payment_id=$1`, mpPaymentID).Scan(&charID)
		if err == sql.ErrNoRows {
			return 0, 0, err
		}
		if err != nil {
			return 0, 0, err
		}
		if err = tx.Commit(); err != nil {
			return 0, 0, err
		}
		return charID, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	if _, err = tx.Exec(`UPDATE characters SET diamonds=diamonds+$1 WHERE id=$2`, p.Diamonds, p.CharacterID); err != nil {
		return 0, 0, err
	}
	if _, err = tx.Exec(`INSERT INTO diamond_log (character_id, amount, reason) VALUES ($1,$2,'pix_mercadopago')`, p.CharacterID, p.Diamonds); err != nil {
		return 0, 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return p.CharacterID, p.Diamonds, nil
}

// ConfirmPixPayment marks payment as paid and grants diamonds.
// In production this is called by the Pix webhook from your bank.
func ConfirmPixPayment(txid string) (charID, diamonds int, err error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Atomic idempotency: only one concurrent confirmer can flip pending -> paid.
	var p PixPaymentRecord
	err = tx.QueryRow(`
		UPDATE pix_payments
		SET status='paid', paid_at=NOW()
		WHERE txid=$1 AND status='pending'
		RETURNING id, character_id, diamonds
	`, txid).Scan(&p.ID, &p.CharacterID, &p.Diamonds)
	if err == sql.ErrNoRows {
		// Already processed (or missing payment). Keep compatibility by returning charID when found.
		err = tx.QueryRow(`SELECT character_id FROM pix_payments WHERE txid=$1`, txid).Scan(&charID)
		if err == sql.ErrNoRows {
			return 0, 0, err
		}
		if err != nil {
			return 0, 0, err
		}
		if err = tx.Commit(); err != nil {
			return 0, 0, err
		}
		return charID, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	if _, err = tx.Exec(`UPDATE characters SET diamonds=diamonds+$1 WHERE id=$2`, p.Diamonds, p.CharacterID); err != nil {
		return 0, 0, err
	}
	if _, err = tx.Exec(`INSERT INTO diamond_log (character_id, amount, reason) VALUES ($1,$2,'pix_payment')`, p.CharacterID, p.Diamonds); err != nil {
		return 0, 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return p.CharacterID, p.Diamonds, nil
}

// GetRecentPixPayments returns the most recent N pix payments (any status).
func GetRecentPixPayments(limit int) ([]PixPaymentRecord, error) {
	rows, err := DB.Query(`
		SELECT id, character_id, package_id, diamonds, amount_brl, txid, status,
		       COALESCE(mp_payment_id, 0), COALESCE(qr_code, ''), COALESCE(qr_code_b64, '')
		FROM pix_payments
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payments []PixPaymentRecord
	for rows.Next() {
		var p PixPaymentRecord
		if err := rows.Scan(&p.ID, &p.CharacterID, &p.PackageID, &p.Diamonds, &p.AmountBRL,
			&p.TxID, &p.Status, &p.MPPaymentID, &p.QRCode, &p.QRCodeB64); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payments, nil
}

// ConfirmPixPaymentByTxID é um alias de ConfirmPixPayment, usado pelo AbacatePay.
func ConfirmPixPaymentByTxID(txID string) (charID, diamonds int, err error) {
	return ConfirmPixPayment(txID)
}
