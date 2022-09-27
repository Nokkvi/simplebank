package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/nokkvi/simplebank/db/mock"
	db "github.com/nokkvi/simplebank/db/sqlc"
	"github.com/nokkvi/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestGetEntry(t *testing.T) {
	entry := randomEntry(account.ID)

	testCases := []struct{
		name string
		entryID int64
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntry(t, recorder.Body, entry)
			},
		},
		{
			name: "NotFound",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(db.Entry{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			entryID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Any()).Times(0)
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
		
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)
			
		
			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()
		
			url := fmt.Sprintf("/entries/%d", tc.entryID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
		
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListEntries(t *testing.T) {
	entries := randomEntries(account.ID, 49)

	testCases := []struct{
		name string
		pageID int64
		pageSize int64
		accountID int64
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "FirstPage",
			pageID: 1,
			pageSize: 5,
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).Times(1).Return(entries[:5], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntries(t, recorder.Body, entries[:5])
				require.Equal(t, len(entries[:5]), 5)
			},
		},
		{
			name: "LastPage",
			pageID: 1,
			pageSize: 5,
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).Times(1).Return(entries[45:], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntries(t, recorder.Body, entries[45:])
				require.Equal(t, len(entries[45:]), 4)
			},
		},
		{
			name: "BadRequestAccountID",
			pageID: 0,
			pageSize: 5,
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestPageID",
			pageID: 0,
			pageSize: 5,
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestPageSize",
			pageID: 1,
			pageSize: 1,
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			pageID: 1,
			pageSize: 5,
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).Times(1).Return([]db.Entry{}, sql.ErrConnDone)
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
		
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)
			
		
			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()
		
			url := fmt.Sprintf("/entries?page_id=%d&page_size=%d&account_id=%d", tc.pageID, tc.pageSize, tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
		
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomEntry(accountID int64) db.Entry {
	return db.Entry{
		ID: util.RandomInt(0, 1000),
		AccountID: accountID,
		Amount: util.RandomMoney(),
		CreatedAt: util.RandomDate(),
	}
}

func randomEntries(accountID int64, size int) []db.Entry {
	var entries []db.Entry

	for i := 0; i < size; i++ {
		entries = append(entries, randomEntry(accountID))
	}

	return entries
}

func requireBodyMatchEntry(t *testing.T, body *bytes.Buffer, entry db.Entry) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotEntry db.Entry
	err = json.Unmarshal(data, &gotEntry)
	require.NoError(t, err)
	require.Equal(t, entry, gotEntry)
}

func requireBodyMatchEntries(t *testing.T, body *bytes.Buffer, entries []db.Entry) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotEntries []db.Entry
	err = json.Unmarshal(data, &gotEntries)
	require.NoError(t, err)
	require.Equal(t, entries, gotEntries)
}