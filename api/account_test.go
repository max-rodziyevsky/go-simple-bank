package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/max-rodziyevsky/go-simple-bank/internal/repo"
	mockrepo "github.com/max-rodziyevsky/go-simple-bank/internal/repo/mock"
	"github.com/max-rodziyevsky/go-simple-bank/util"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetAccountAPI
func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(mockStore *mockrepo.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(repo.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(repo.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockrepo.NewMockStore(ctrl)
			tc.buildStubs(mockStore)

			server := NewServer(mockStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}

func TestCreateAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		arg           gin.H
		buildStubs    func(mockStore *mockrepo.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			arg: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				arg := repo.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
				}

				mockStore.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "InvalidOwner",
			arg: gin.H{
				"owner":    "",
				"currency": account.Currency,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			arg: gin.H{
				"owner":    account.Owner,
				"currency": "",
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			arg: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(repo.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockrepo.NewMockStore(ctrl)
			tc.buildStubs(mockStore)

			server := NewServer(mockStore)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.arg)
			require.NoError(t, err)

			url := fmt.Sprint("/accounts")
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccounts(t *testing.T) {
	n := 5
	accounts := make([]repo.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount()
	}

	type Query struct {
		PageID   int32
		PageSize int32
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(mockStore *mockrepo.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				PageID:   1,
				PageSize: int32(n),
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				arg := repo.ListAccountsParams{
					Offset: 0,
					Limit:  int32(n),
				}

				mockStore.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				PageID:   0,
				PageSize: int32(n),
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			query: Query{
				PageID:   1,
				PageSize: 0,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			query: Query{
				PageID:   1,
				PageSize: int32(n),
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]repo.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockrepo.NewMockStore(ctrl)
			tc.buildStubs(mockStore)

			server := NewServer(mockStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprint("/accounts")
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			q.Add("page_id", fmt.Sprint(tc.query.PageID))
			q.Add("page_size", fmt.Sprint(tc.query.PageSize))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	account := randomAccount()
	newBalance := util.RandomMoney()

	updatedAccount := account
	updatedAccount.Balance = newBalance

	type req updateAccountRequest

	testCases := []struct {
		name          string
		req           req
		buildStubs    func(mockStore *mockrepo.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			req: req{
				ID:      account.ID,
				Balance: newBalance,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				arg := repo.UpdateAccountParams{
					ID:      account.ID,
					Balance: newBalance,
				}
				mockStore.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedAccount, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccountUpdate(t, recorder.Body, account)
			},
		},
		{
			name: "InternalServerError",
			req: req{
				ID:      account.ID,
				Balance: newBalance,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(repo.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "NotFound",
			req: req{
				ID:      account.ID,
				Balance: newBalance,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(repo.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			req: req{
				ID:      0,
				Balance: newBalance,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidBalance",
			req: req{
				ID:      account.ID,
				Balance: -1,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockrepo.NewMockStore(ctrl)
			tc.buildStubs(mockStore)

			server := NewServer(mockStore)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.req)
			require.NoError(t, err)

			url := fmt.Sprint("/accounts")
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))

			server.router.ServeHTTP(recorder, request)
			requireBodyMatchAccountUpdate(t, recorder.Body, account)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name       string
		accountID  int64
		buildStubs func(mockStore *mockrepo.MockStore)
		statusCode int
	}{
		{
			name:      "Deleted",
			accountID: account.ID,
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1)
			},
			statusCode: http.StatusNoContent,
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			statusCode: http.StatusInternalServerError,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockrepo.NewMockStore(ctrl)
			tc.buildStubs(mockStore)

			server := NewServer(mockStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			require.Equal(t, tc.statusCode, recorder.Code)
		})
	}
}

// utility methods:

// getAccountAPI this is an example of a single test with explanation on each step. TestGetAccountAPI - this is more advance approach to unit testing
func getAccountAPI(t *testing.T) {
	// First we need to an account to work with
	account := randomAccount()

	// Then we need a mock database instance, we call NewMockStore method that generated by mock package
	// This method receive gomock.Controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := mockrepo.NewMockStore(ctrl)

	// Then we can build stubs (imitate service or handler in my case behavior)
	mockStore.EXPECT().
		GetAccount(gomock.Any(), gomock.Eq(account.ID)).
		Times(1).
		Return(account, nil)

	// Then we should start test server and send request
	server := NewServer(mockStore)
	// To test http server we don't have to start it
	// We can use the feature of httptest package or record http response
	recorder := httptest.NewRecorder()

	// form url
	url := fmt.Sprintf("/accounts/%d", account.ID)
	// Preparing request - it will return request object and an error
	request, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	// So in this part we will serve http server - send our prepared request and write down to recorder response
	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	// then we need to compare recorder body
	requireBodyMatchAccount(t, recorder.Body, account)
}

func randomAccount() repo.Account {
	return repo.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomString(6),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account repo.Account) {
	response, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount repo.Account
	err = json.Unmarshal(response, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []repo.Account) {
	response, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []repo.Account
	err = json.Unmarshal(response, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}

func requireBodyMatchAccountUpdate(t *testing.T, body *bytes.Buffer, account repo.Account) {
	response, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUpdatedAccount repo.Account
	err = json.Unmarshal(response, &gotUpdatedAccount)
	require.NoError(t, err)
	require.NotEqual(t, account.Balance, gotUpdatedAccount.Balance)
}
