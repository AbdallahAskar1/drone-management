package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"drone-management/internal/database"
	"drone-management/internal/domain"
	"drone-management/internal/handler"
	"drone-management/internal/repo"
	"drone-management/internal/service"
	"drone-management/internal/utils"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestApp(t *testing.T) (*echo.Echo, *utils.JWTSigner) {
	dbURL := "postgres://postgres:postgres@localhost:5432/drones_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Skip("Postgres not available for integration tests")
	}

	database.Migrate(db)
	db.Exec("TRUNCATE principals, drones, orders, jobs, order_events CASCADE")

	clock := utils.RealClock{}
	signer := utils.NewJWTSigner("test-secret", time.Hour)

	pRepo := repo.NewPrincipalRepo(db)
	dRepo := repo.NewDroneRepo(db)
	oRepo := repo.NewOrderRepo(db)
	jRepo := repo.NewJobRepo(db)
	eRepo := repo.NewEventRepo(db)

	authSvc := service.NewAuthService(pRepo, dRepo, signer, clock)
	orderSvc := service.NewOrderService(db, oRepo, jRepo, eRepo, clock)
	droneSvc := service.NewDroneService(db, dRepo, oRepo, jRepo, eRepo, pRepo, clock, 10)
	adminSvc := service.NewAdminService(db, oRepo, dRepo, eRepo, clock)

	h := handler.Handlers{
		Auth:    handler.NewAuthHandler(authSvc),
		Enduser: handler.NewEnduserHandler(orderSvc),
		Drone:   handler.NewDroneHandler(droneSvc),
		Admin:   handler.NewAdminHandler(adminSvc, droneSvc),
	}

	e := echo.New()
	handler.Register(e, signer, h)
	return e, signer
}

func request(e *echo.Echo, method, path string, body any, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func uintToStr(n uint) string {
	return utils.UintToStr(n)
}

func TestOrderLifecycle(t *testing.T) {
	e, signer := setupTestApp(t)

	userTok, _ := signer.Issue(2, "User", domain.RoleEnduser, time.Now())
	droneTok, _ := signer.Issue(3, "Drone1", domain.RoleDrone, time.Now())

	subReq := handler.SubmitOrderRequest{
		Origin:      handler.LatLng{Lat: 0, Lng: 0},
		Destination: handler.LatLng{Lat: 1, Lng: 1},
	}
	rec := request(e, http.MethodPost, "/orders", subReq, userTok)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var orderRes handler.OrderResponse
	json.Unmarshal(rec.Body.Bytes(), &orderRes)
	orderID := orderRes.ID

	request(e, http.MethodPost, "/drone/jobs/1/reserve", nil, droneTok)
	request(e, http.MethodPost, "/drone/orders/"+uintToStr(orderID)+"/pickup", nil, droneTok)
	request(e, http.MethodPost, "/drone/orders/"+uintToStr(orderID)+"/delivered", nil, droneTok)

	rec = request(e, http.MethodGet, "/orders/"+uintToStr(orderID), nil, userTok)
	json.Unmarshal(rec.Body.Bytes(), &orderRes)
	assert.Equal(t, domain.OrderStatusDelivered, orderRes.Status)
}

func TestWithdrawOrder(t *testing.T) {
	e, signer := setupTestApp(t)
	userTok, _ := signer.Issue(2, "User", domain.RoleEnduser, time.Now())

	subReq := handler.SubmitOrderRequest{
		Origin:      handler.LatLng{Lat: 0, Lng: 0},
		Destination: handler.LatLng{Lat: 1, Lng: 1},
	}
	rec := request(e, http.MethodPost, "/orders", subReq, userTok)
	var orderRes handler.OrderResponse
	json.Unmarshal(rec.Body.Bytes(), &orderRes)
	orderID := orderRes.ID

	rec = request(e, http.MethodPost, "/orders/"+uintToStr(orderID)+"/withdraw", nil, userTok)
	assert.Equal(t, http.StatusOK, rec.Code)

	rec = request(e, http.MethodGet, "/orders/"+uintToStr(orderID), nil, userTok)
	json.Unmarshal(rec.Body.Bytes(), &orderRes)
	assert.Equal(t, domain.OrderStatusWithdrawn, orderRes.Status)
}

func TestRBAC(t *testing.T) {
	e, signer := setupTestApp(t)
	userTok, _ := signer.Issue(2, "User", domain.RoleEnduser, time.Now())
	droneTok, _ := signer.Issue(3, "Drone", domain.RoleDrone, time.Now())

	rec := request(e, http.MethodPost, "/orders", nil, droneTok)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	rec = request(e, http.MethodGet, "/drone/jobs", nil, userTok)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandoff(t *testing.T) {
	e, signer := setupTestApp(t)
	userTok, _ := signer.Issue(2, "User", domain.RoleEnduser, time.Now())
	drone1Tok, _ := signer.Issue(3, "Drone1", domain.RoleDrone, time.Now())
	drone2Tok, _ := signer.Issue(4, "Drone2", domain.RoleDrone, time.Now())

	subReq := handler.SubmitOrderRequest{
		Origin:      handler.LatLng{Lat: 0, Lng: 0},
		Destination: handler.LatLng{Lat: 1, Lng: 1},
	}
	rec := request(e, http.MethodPost, "/orders", subReq, userTok)
	var orderRes handler.OrderResponse
	json.Unmarshal(rec.Body.Bytes(), &orderRes)
	orderID := orderRes.ID

	request(e, http.MethodPost, "/drone/jobs/1/reserve", nil, drone1Tok)
	request(e, http.MethodPost, "/drone/orders/"+uintToStr(orderID)+"/pickup", nil, drone1Tok)

	rec = request(e, http.MethodPost, "/drone/self/broken", nil, drone1Tok)
	assert.Equal(t, http.StatusOK, rec.Code)

	rec = request(e, http.MethodGet, "/orders/"+uintToStr(orderID), nil, userTok)
	json.Unmarshal(rec.Body.Bytes(), &orderRes)
	assert.Equal(t, domain.OrderStatusHandoffRequired, orderRes.Status)

	rec = request(e, http.MethodGet, "/drone/jobs", nil, drone2Tok)
	var jobsResp struct {
		Jobs []*domain.Job `json:"jobs"`
	}
	json.Unmarshal(rec.Body.Bytes(), &jobsResp)
	require.NotEmpty(t, jobsResp.Jobs)
	handoffJob := jobsResp.Jobs[0]

	request(e, http.MethodPost, "/drone/jobs/"+uintToStr(handoffJob.ID)+"/reserve", nil, drone2Tok)
	request(e, http.MethodPost, "/drone/orders/"+uintToStr(orderID)+"/pickup", nil, drone2Tok)
	request(e, http.MethodPost, "/drone/orders/"+uintToStr(orderID)+"/delivered", nil, drone2Tok)

	rec = request(e, http.MethodGet, "/orders/"+uintToStr(orderID), nil, userTok)
	json.Unmarshal(rec.Body.Bytes(), &orderRes)
	assert.Equal(t, domain.OrderStatusDelivered, orderRes.Status)
}

func TestConcurrency(t *testing.T) {
	e, signer := setupTestApp(t)
	drone1Tok, _ := signer.Issue(3, "Drone1", domain.RoleDrone, time.Now())
	drone2Tok, _ := signer.Issue(4, "Drone2", domain.RoleDrone, time.Now())

	var wg sync.WaitGroup
	wg.Add(2)
	results := make(chan int, 2)

	go func() {
		defer wg.Done()
		rec := request(e, http.MethodPost, "/drone/jobs/1/reserve", nil, drone1Tok)
		results <- rec.Code
	}()

	go func() {
		defer wg.Done()
		rec := request(e, http.MethodPost, "/drone/jobs/1/reserve", nil, drone2Tok)
		results <- rec.Code
	}()

	wg.Wait()
	close(results)

	codes := []int{}
	for c := range results {
		codes = append(codes, c)
	}
	assert.Contains(t, codes, http.StatusOK)
}
