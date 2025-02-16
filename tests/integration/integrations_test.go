package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	authHandler "Merch_store-Avito_test_task/internal/pkg/auth/delivery/http"
	authRepo "Merch_store-Avito_test_task/internal/pkg/auth/repository"
	authUsecase "Merch_store-Avito_test_task/internal/pkg/auth/usecase"
	"Merch_store-Avito_test_task/internal/pkg/jwt"
	"Merch_store-Avito_test_task/internal/pkg/middleware"
	paymentsHandler "Merch_store-Avito_test_task/internal/pkg/payments/delivery/http"
	paymentsRepo "Merch_store-Avito_test_task/internal/pkg/payments/repository"
	paymentsUsecase "Merch_store-Avito_test_task/internal/pkg/payments/usecase"
)

type IntegrationTestSuite struct {
	suite.Suite
	db              *sql.DB
	router          *mux.Router
	server          *httptest.Server
	paymentsHandler *paymentsHandler.PaymentsHandler
	authHandler     *authHandler.AuthHandler
	jwtHandler      jwt.JWTInterface
	logger          *slog.Logger
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Подключение к тестовой БД
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		5433,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
	))
	require.NoError(s.T(), err)
	s.db = db

	// Инициализация JWT
	s.jwtHandler = jwt.NewJTW("test-secret", s.logger)

	// Инициализация слоев
	// Auth
	authRepo := authRepo.NewAuthRepositoryImpl(s.db)
	authUc := authUsecase.NewAuthUsecase(authRepo)
	s.authHandler = authHandler.NewAuthHandler(authUc, s.logger, s.jwtHandler)

	// Payments
	paymentsRepo := paymentsRepo.NewAuthRepositoryImpl(s.db)
	paymentsUc := paymentsUsecase.NewPaymentsUsecase(paymentsRepo)
	s.paymentsHandler = paymentsHandler.NewPaymentsHandler(paymentsUc, s.logger)

	// Router setup
	s.router = mux.NewRouter().PathPrefix("/api").Subrouter()
	s.setupRoutes()

	s.server = httptest.NewServer(s.router)

	// Создание тестовых таблиц
	err = s.createTestTables()
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) setupRoutes() {
	s.router.HandleFunc("/auth", s.authHandler.Login).Methods(http.MethodPost)
	s.router.Handle("/sendCoin", middleware.AuthMiddleware(s.jwtHandler,
		http.HandlerFunc(s.paymentsHandler.SendCoins), s.logger)).Methods(http.MethodPost)
	s.router.Handle("/buy/{item}", middleware.AuthMiddleware(s.jwtHandler,
		http.HandlerFunc(s.paymentsHandler.BuyItem), s.logger)).Methods(http.MethodGet)
}

func (s *IntegrationTestSuite) createTestTables() error {
	queries := []string{
		`DROP TABLE IF EXISTS "transaction" CASCADE`,
		`DROP TABLE IF EXISTS "purchase" CASCADE`,
		`DROP TABLE IF EXISTS "product" CASCADE`,
		`DROP TABLE IF EXISTS "user" CASCADE`,
		`CREATE TABLE "user" (
            id SERIAL PRIMARY KEY,
            username VARCHAR(255) UNIQUE NOT NULL,
            coins INTEGER NOT NULL CHECK (coins >= 0)
        )`,
		`CREATE TABLE "product" (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            price INTEGER NOT NULL
        )`,
		`CREATE TABLE "purchase" (
            id SERIAL PRIMARY KEY,
            user_id INTEGER REFERENCES "user"(id),
            product_id INTEGER REFERENCES "product"(id),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE "transaction" (
            id SERIAL PRIMARY KEY,
            amount INTEGER NOT NULL,
            from_user_id INTEGER REFERENCES "user"(id),
            to_user_id INTEGER REFERENCES "user"(id),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.server.Close()
	s.db.Close()
}

func (s *IntegrationTestSuite) TearDownTest() {
	// Очистка таблиц после каждого теста
	tables := []string{"transaction", "purchase", "product", "user"}
	for _, table := range tables {
		_, err := s.db.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, table))
		require.NoError(s.T(), err)
	}
}

// Тест успешной покупки товара
func (s *IntegrationTestSuite) TestSuccessfulPurchase() {
	// Создание тестового пользователя
	userID := s.createTestUser("testuser", 1000)
	token := s.generateTestToken(userID, "testuser")

	// Создание тестового товара
	itemID := s.createTestProduct("Test T-Shirt", 500)

	// Выполнение запроса на покупку
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/buy/%d", itemID), nil)
	req.Header.Set("Access-Token", token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Проверки
	s.Equal(http.StatusOK, w.Code)

	// Проверка баланса пользователя
	var balance int
	err := s.db.QueryRow(`SELECT coins FROM "user" WHERE id = $1`, userID).Scan(&balance)
	s.NoError(err)
	s.Equal(500, balance)

	// Проверка создания записи о покупке
	var purchaseCount int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM "purchase" WHERE user_id = $1 AND product_id = $2`,
		userID, itemID).Scan(&purchaseCount)
	s.NoError(err)
	s.Equal(1, purchaseCount)
}

// Тест покупки товара c плохим запросом
func (s *IntegrationTestSuite) TestPurchaseBadRequest() {
	// Создание тестового пользователя
	userID := s.createTestUser("testuser", 1000)
	token := s.generateTestToken(userID, "testuser")

	// Создание тестового товара
	itemID := s.createTestProduct("Test T-Shirt", 500)

	// Выполнение запроса на покупку
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/buy/%s", "err"), nil)
	req.Header.Set("Access-Token", token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Проверки
	s.Equal(http.StatusBadRequest, w.Code)

	// Проверка баланса пользователя
	var balance int
	err := s.db.QueryRow(`SELECT coins FROM "user" WHERE id = $1`, userID).Scan(&balance)
	s.NoError(err)
	s.Equal(1000, balance)

	// Проверка создания записи о покупке
	var purchaseCount int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM "purchase" WHERE user_id = $1 AND product_id = $2`,
		userID, itemID).Scan(&purchaseCount)
	s.NoError(err)
	s.Equal(0, purchaseCount)
}

// Тест покупки несуществующего товара
func (s *IntegrationTestSuite) TestPurchaseNonExistingProduct() {
	// Создание тестового пользователя
	userID := s.createTestUser("testuser", 1000)
	token := s.generateTestToken(userID, "testuser")

	itemID := 100000

	// Выполнение запроса на покупку
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/buy/%d", itemID), nil)
	req.Header.Set("Access-Token", token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Проверки
	s.Equal(http.StatusInternalServerError, w.Code)

	// Проверка баланса пользователя
	var balance int
	err := s.db.QueryRow(`SELECT coins FROM "user" WHERE id = $1`, userID).Scan(&balance)
	s.NoError(err)
	s.Equal(1000, balance)

	// Проверка создания записи о покупке
	var purchaseCount int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM "purchase" WHERE user_id = $1 AND product_id = $2`,
		userID, itemID).Scan(&purchaseCount)
	s.NoError(err)
	s.Equal(0, purchaseCount)
}

// Тест покупки при недостаточном балансе
func (s *IntegrationTestSuite) TestInsufficientFundsPurchase() {
	userID := s.createTestUser("pooruser", 100)
	token := s.generateTestToken(userID, "pooruser")
	itemID := s.createTestProduct("Expensive Item", 500)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/buy/%d", itemID), nil)
	req.Header.Set("Access-Token", token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusForbidden, w.Code)

	var balance int
	err := s.db.QueryRow(`SELECT coins FROM "user" WHERE id = $1`, userID).Scan(&balance)
	s.NoError(err)
	s.Equal(100, balance)
}

// Тест успешной передачи монет
func (s *IntegrationTestSuite) TestSuccessfulCoinTransfer() {
	senderID := s.createTestUser("sender", 1000)
	receiverID := s.createTestUser("receiver", 0)
	token := s.generateTestToken(senderID, "sender")

	sendData := map[string]interface{}{
		"toUser": "receiver",
		"amount": 500,
	}
	body, err := json.Marshal(sendData)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin",
		bytes.NewBuffer(body))
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Проверка балансов
	var senderBalance, receiverBalance int
	err = s.db.QueryRow(`SELECT coins FROM "user" WHERE id = $1`, senderID).Scan(&senderBalance)
	s.NoError(err)
	s.Equal(500, senderBalance)

	err = s.db.QueryRow(`SELECT coins FROM "user" WHERE id = $1`, receiverID).Scan(&receiverBalance)
	s.NoError(err)
	s.Equal(500, receiverBalance)
}

// Тест передачи монет несуществующему пользователю
func (s *IntegrationTestSuite) TestTransferToNonExistentUser() {
	senderID := s.createTestUser("sender", 1000)
	token := s.generateTestToken(senderID, "sender")

	sendData := map[string]interface{}{
		"toUser": "nonexistent",
		"amount": 500,
	}
	body, err := json.Marshal(sendData)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin",
		bytes.NewBuffer(body))
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)

	// Проверка что баланс отправителя не изменился
	var senderBalance int
	err = s.db.QueryRow(`SELECT coins FROM "user" WHERE id = $1`, senderID).Scan(&senderBalance)
	s.NoError(err)
	s.Equal(1000, senderBalance)
}

// Тест передачи отрицательного количества монет
func (s *IntegrationTestSuite) TestCoinTransferBadRequest() {
	senderID := s.createTestUser("sender", 1000)
	receiverID := s.createTestUser("receiver", 0)
	token := s.generateTestToken(senderID, "sender")

	sendData := map[string]interface{}{
		"toUser": "receiver",
		"amount": -100,
	}
	body, err := json.Marshal(sendData)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin",
		bytes.NewBuffer(body))
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	// Проверка балансов
	var senderBalance, receiverBalance int
	err = s.db.QueryRow(`SELECT coins FROM "user" WHERE id = $1`, senderID).Scan(&senderBalance)
	s.NoError(err)
	s.Equal(1000, senderBalance)

	err = s.db.QueryRow(`SELECT coins FROM "user" WHERE id = $1`, receiverID).Scan(&receiverBalance)
	s.NoError(err)
	s.Equal(0, receiverBalance)
}

func (s *IntegrationTestSuite) createTestUser(username string, coins int) uint {
	var id uint
	err := s.db.QueryRow(
		`INSERT INTO "user" (username, coins) VALUES ($1, $2) RETURNING id`,
		username, coins,
	).Scan(&id)
	s.NoError(err)
	return id
}

func (s *IntegrationTestSuite) createTestProduct(name string, price int) uint {
	var id uint
	err := s.db.QueryRow(
		`INSERT INTO "product" (name, price) VALUES ($1, $2) RETURNING id`,
		name, price,
	).Scan(&id)
	s.NoError(err)
	return id
}

func (s *IntegrationTestSuite) generateTestToken(userID uint, username string) string {
	token, err := s.jwtHandler.GenerateToken(userID, username)
	s.NoError(err)
	return token
}
