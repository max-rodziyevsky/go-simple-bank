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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTransfer(t *testing.T) {
	account1 := randomAccount()
	account2 := randomAccount()
	account3 := randomAccount()

	account1.Currency = util.USD
	account2.Currency = util.USD
	account3.Currency = util.EUR

	amount := util.RandomMoney()

	testCases := []struct {
		name          string
		req           gin.H
		buildStubs    func(mockStore *mockrepo.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			req: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)
				mockStore.EXPECT().GetAccount(gomock.Any(), account2.ID).Times(1).Return(account2, nil)

				arg := repo.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}

				mockStore.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).
					Times(1)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InvalidAmount",
			req: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          0,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(0).Return(account1, nil)
				mockStore.EXPECT().GetAccount(gomock.Any(), account2.ID).Times(0).Return(account2, nil)

				mockStore.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidFromAccountID",
			req: gin.H{
				"from_account_id": 0,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(0).Return(account1, nil)
				mockStore.EXPECT().GetAccount(gomock.Any(), account2.ID).Times(0).Return(account2, nil)

				mockStore.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			req: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(repo.Account{}, sql.ErrNoRows)
				mockStore.EXPECT().GetAccount(gomock.Any(), account2.ID).Times(0)
				mockStore.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "ToAccountNotFound",
			req: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)
				mockStore.EXPECT().GetAccount(gomock.Any(), account2.ID).Times(1).Return(repo.Account{}, sql.ErrNoRows)
				mockStore.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			req: gin.H{
				"from_account_id": account3.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account3.ID)).Times(1).Return(account3, nil)
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				mockStore.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			req: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account3.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account3.ID)).Times(1).Return(account3, nil)
				mockStore.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			req: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        "XYZ",
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(0)
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(0)
				mockStore.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidAmount",
			req: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          -amount,
				"currency":        "XYZ",
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(0)
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(0)
				mockStore.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "GetAccountInternalError",
			req: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(repo.Account{}, sql.ErrConnDone)
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(0)
				mockStore.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "TransferTxError",
			req: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockrepo.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(account2, nil)
				mockStore.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(repo.TransferTxResult{}, sql.ErrTxDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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

			data, err := json.Marshal(tc.req)
			require.NoError(t, err)

			url := fmt.Sprint("/transfers")
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})

	}
}
