package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	paymentdomain "Microservice-Payments-Service/payments/domain"
)

func TestRespondErrorMapsStatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "external provider", err: paymentdomain.ErrExternalProvider, wantStatus: http.StatusBadGateway},
		{name: "duplicate webhook", err: paymentdomain.ErrDuplicateWebhook, wantStatus: http.StatusConflict},
		{name: "invalid uuid", err: paymentdomain.ErrInvalidUUID, wantStatus: http.StatusBadRequest},
		{name: "postgres unique violation", err: &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"}, wantStatus: http.StatusConflict},
		{name: "postgres foreign key violation", err: &pgconn.PgError{Code: "23503", Message: "insert or update violates foreign key constraint"}, wantStatus: http.StatusConflict},
		{name: "postgres invalid text", err: &pgconn.PgError{Code: "22P02", Message: "invalid input syntax for type uuid"}, wantStatus: http.StatusBadRequest},
		{name: "unknown", err: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	gin.SetMode(gin.TestMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			RespondError(ctx, tc.err)

			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d body=%s", recorder.Code, tc.wantStatus, recorder.Body.String())
			}
		})
	}
}
