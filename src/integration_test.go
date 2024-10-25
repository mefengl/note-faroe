package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"faroe/otp"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEndpointResponses(t *testing.T) {
	t.Parallel()

	t.Run("post /users", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/users")

		db := initializeTestDB(t)
		defer db.Close()

		user1 := User{
			Id:             "1",
			CreatedAt:      time.Unix(time.Now().Unix(), 0),
			Email:          "user1@example.com",
			PasswordHash:   "HASH1",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"email","password":"12345678"}`))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidData)

		r = httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"user1@example.com","password":"12345678"}`))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorEmailAlreadyUsed)

		r = httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"USER1@EXAMPLE.COM","password":"12345678"}`))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorEmailAlreadyUsed)

		r = httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"user2@example.com","password":"1234"}`))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorWeakPassword)

		r = httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"user2@example.com","password":"12345678"}`))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorWeakPassword)

		r = httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"user2@example.com","password":"super_secure_password"}`))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, userJSONKeys)
	})

	t.Run("get /users", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "GET", "/users")

		t.Run("sort order", func(t *testing.T) {
			t.Parallel()
			db := initializeTestDB(t)
			defer db.Close()

			now := time.Unix(time.Now().Unix(), 0)

			user1 := User{
				Id:             "1",
				Email:          "a@example.com",
				CreatedAt:      time.Unix(now.Add(1*time.Second).Unix(), 0),
				PasswordHash:   "HASH1",
				RecoveryCode:   "CODE1",
				TOTPRegistered: false,
			}
			err := insertUser(db, context.Background(), &user1)
			if err != nil {
				t.Fatal(err)
			}

			user2 := User{
				Id:             "2",
				Email:          "c@example.com",
				CreatedAt:      now,
				PasswordHash:   "HASH2",
				RecoveryCode:   "CODE2",
				TOTPRegistered: false,
			}
			err = insertUser(db, context.Background(), &user2)
			if err != nil {
				t.Fatal(err)
			}

			user3 := User{
				Id:           "3",
				Email:        "b@example.com",
				CreatedAt:    time.Unix(now.Add(2*time.Second).Unix(), 0),
				PasswordHash: "HASH3",
				RecoveryCode: "CODE3",
			}
			err = insertUser(db, context.Background(), &user3)
			if err != nil {
				t.Fatal(err)
			}

			env := createEnvironment(db, nil)
			app := CreateApp(env)

			testCases := []struct {
				SortBy    string
				SortOrder string
				Expected  []User
			}{
				{"created_at", "ascending", []User{user2, user1, user3}},
				{"created_at", "descending", []User{user3, user1, user2}},
				{"email", "ascending", []User{user1, user3, user2}},
				{"email", "descending", []User{user2, user3, user1}},
				{"id", "ascending", []User{user1, user2, user3}},
				{"id", "descending", []User{user3, user2, user1}},
				{"", "", []User{user2, user1, user3}},
			}

			for _, testCase := range testCases {
				values := url.Values{}
				values.Set("sort_by", testCase.SortBy)
				values.Set("sort_order", testCase.SortOrder)
				url := "/users?" + values.Encode()
				r := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()
				app.ServeHTTP(w, r)
				res := w.Result()
				assert.Equal(t, 200, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				var result []UserJSON
				err = json.Unmarshal(body, &result)
				if err != nil {
					t.Fatal(err)
				}

				var expected []UserJSON
				for _, expectedItem := range testCase.Expected {
					var item UserJSON
					err = json.Unmarshal([]byte(expectedItem.EncodeToJSON()), &item)
					if err != nil {
						t.Fatal(err)
					}
					expected = append(expected, item)
				}

				assert.Equal(t, expected, result)
			}
		})

		t.Run("pagination", func(t *testing.T) {
			t.Parallel()
			db := initializeTestDB(t)
			defer db.Close()

			now := time.Unix(time.Now().Unix(), 0)

			for i := 0; i < 30; i++ {
				user := User{
					Id:             strconv.Itoa(i + 1),
					Email:          fmt.Sprintf("user%d@example.com", i+1),
					CreatedAt:      time.Unix(now.Add(time.Duration(i*int(time.Second))).Unix(), 0),
					PasswordHash:   "HASH",
					RecoveryCode:   "CODE",
					TOTPRegistered: false,
				}
				err := insertUser(db, context.Background(), &user)
				if err != nil {
					t.Fatal(err)
				}
			}

			env := createEnvironment(db, nil)
			app := CreateApp(env)

			testCases := []struct {
				PerPage            string
				Page               string
				ExpectedIdStart    int
				ExpectedIdEnd      int
				ExpectedTotalPages int
			}{
				{"10", "2", 11, 21, 3},
				{"20", "2", 21, 31, 2},
				{"30", "2", 31, 31, 1},
				{"", "2", 21, 31, 2},
				{"a", "2", 21, 31, 2},
				{"-1", "2", 21, 31, 2},
				{"0", "2", 21, 31, 2},

				{"10", "1", 1, 11, 3},
				{"10", "2", 11, 21, 3},
				{"10", "3", 21, 31, 3},
				{"10", "4", 31, 31, 3},
				{"10", "0", 1, 11, 3},
				{"10", "-1", 1, 11, 3},
				{"10", "", 1, 11, 3},
				{"10", "a", 1, 11, 3},

				{"a", "a", 1, 21, 2},
				{"", "", 1, 21, 2},
			}

			for _, testCase := range testCases {
				values := url.Values{}
				values.Set("per_page", testCase.PerPage)
				values.Set("page", testCase.Page)
				values.Set("created_at", "id")
				url := "/users?" + values.Encode()
				r := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()
				app.ServeHTTP(w, r)
				res := w.Result()
				assert.Equal(t, 200, res.StatusCode)

				assert.Equal(t, "30", res.Header.Get("X-Pagination-Total"))
				assert.Equal(t, strconv.Itoa(testCase.ExpectedTotalPages), res.Header.Get("X-Pagination-Total-Pages"))

				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				var result []UserJSON
				err = json.Unmarshal(body, &result)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, testCase.ExpectedIdEnd-testCase.ExpectedIdStart, len(result), fmt.Sprintf(`count: %s, page: %s`, testCase.PerPage, testCase.Page))

				for i := testCase.ExpectedIdStart; i < testCase.ExpectedIdEnd; i++ {
					assert.Equal(t, result[i-testCase.ExpectedIdStart].Id, strconv.Itoa(i), fmt.Sprintf(`count: %s, page: %s`, testCase.PerPage, testCase.Page))
				}
			}

		})
		t.Run("query", func(t *testing.T) {
			t.Parallel()
			db := initializeTestDB(t)
			defer db.Close()

			now := time.Unix(time.Now().Unix(), 0)

			user1 := User{
				Id:             "1",
				Email:          "user1@example.com",
				CreatedAt:      time.Unix(now.Unix(), 0),
				PasswordHash:   "HASH1",
				RecoveryCode:   "CODE1",
				TOTPRegistered: false,
			}
			err := insertUser(db, context.Background(), &user1)
			if err != nil {
				t.Fatal(err)
			}

			user2 := User{
				Id:             "2",
				Email:          "user2@example.com",
				CreatedAt:      time.Unix(now.Add(1*time.Second).Unix(), 0),
				PasswordHash:   "HASH2",
				RecoveryCode:   "CODE2",
				TOTPRegistered: false,
			}
			err = insertUser(db, context.Background(), &user2)
			if err != nil {
				t.Fatal(err)
			}

			user3 := User{
				Id:           "3",
				Email:        "user3@mail.com",
				CreatedAt:    time.Unix(now.Add(2*time.Second).Unix(), 0),
				PasswordHash: "HASH3",
				RecoveryCode: "CODE3",
			}
			err = insertUser(db, context.Background(), &user3)
			if err != nil {
				t.Fatal(err)
			}

			env := createEnvironment(db, nil)
			app := CreateApp(env)

			values := url.Values{}
			values.Set("email_query", "@example.com")
			requestURL := "/users?" + values.Encode()
			r := httptest.NewRequest("GET", requestURL, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, r)
			res := w.Result()
			assert.Equal(t, 200, res.StatusCode)
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			var result []UserJSON
			err = json.Unmarshal(body, &result)
			if err != nil {
				t.Fatal(err)
			}

			var expected []UserJSON

			var expected1 UserJSON
			err = json.Unmarshal([]byte(user1.EncodeToJSON()), &expected1)
			if err != nil {
				t.Fatal(err)
			}
			expected = append(expected, expected1)

			var expected2 UserJSON
			err = json.Unmarshal([]byte(user2.EncodeToJSON()), &expected2)
			if err != nil {
				t.Fatal(err)
			}
			expected = append(expected, expected2)
			assert.Equal(t, expected, result)

			values = url.Values{}
			values.Set("email_query", "")
			requestURL = "/users?" + values.Encode()
			r = httptest.NewRequest("GET", requestURL, nil)
			w = httptest.NewRecorder()
			app.ServeHTTP(w, r)
			res = w.Result()
			assert.Equal(t, 200, res.StatusCode)
			body, err = io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(body, &result)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, 3, len(result))
		})

		t.Run("pagination with query", func(t *testing.T) {
			t.Parallel()
			db := initializeTestDB(t)
			defer db.Close()

			now := time.Unix(time.Now().Unix(), 0)

			for i := 0; i < 25; i++ {
				user := User{
					Id:             strconv.Itoa(i + 1),
					Email:          fmt.Sprintf("user%d@mail.com", i+1),
					CreatedAt:      time.Unix(now.Add(time.Duration(i*int(time.Second))).Unix(), 0),
					PasswordHash:   "HASH",
					RecoveryCode:   "CODE",
					TOTPRegistered: false,
				}
				err := insertUser(db, context.Background(), &user)
				if err != nil {
					t.Fatal(err)
				}
			}

			for i := 25; i < 30; i++ {
				user := User{
					Id:             strconv.Itoa(i + 1),
					Email:          fmt.Sprintf("user%d@example.com", i+1),
					CreatedAt:      time.Unix(now.Add(time.Duration(i*int(time.Second))).Unix(), 0),
					PasswordHash:   "HASH",
					RecoveryCode:   "CODE",
					TOTPRegistered: false,
				}
				err := insertUser(db, context.Background(), &user)
				if err != nil {
					t.Fatal(err)
				}
			}

			env := createEnvironment(db, nil)
			app := CreateApp(env)

			values := url.Values{}
			values.Set("email_query", "@example.com")
			url := "/users?" + values.Encode()
			r := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, r)
			res := w.Result()
			assert.Equal(t, 200, res.StatusCode)
			assert.Equal(t, "5", res.Header.Get("X-Pagination-Total"))
			assert.Equal(t, "1", res.Header.Get("X-Pagination-Total-Pages"))
		})
	})

	t.Run("get /users/userid", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "GET", "/users/1")

		db := initializeTestDB(t)
		defer db.Close()

		user1 := User{
			Id:             "1",
			CreatedAt:      time.Unix(time.Now().Unix(), 0),
			Email:          "user1@example.com",
			PasswordHash:   "HASH1",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("GET", "/users/2", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/users/1", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		var result UserJSON
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatal(err)
		}
		var expected UserJSON
		err = json.Unmarshal([]byte(user1.EncodeToJSON()), &expected)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, result)
	})

	t.Run("delete /users/userid", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "DELETE", "/users/1")

		db := initializeTestDB(t)
		defer db.Close()

		user1 := User{
			Id:             "1",
			CreatedAt:      time.Unix(time.Now().Unix(), 0),
			Email:          "user1@example.com",
			PasswordHash:   "HASH1",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("DELETE", "/users/2", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/users/1", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("post /users/userid/update-password", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/users/1/update-password")

		db := initializeTestDB(t)
		defer db.Close()

		user1 := User{
			Id:             "1",
			CreatedAt:      time.Unix(time.Now().Unix(), 0),
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("POST", "/users/2/update-password", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		data := `{"password":"invalid","new_password":"1234"}`
		r = httptest.NewRequest("POST", "/users/1/update-password", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorWeakPassword)

		data = `{"password":"invalid","new_password":"12345678"}`
		r = httptest.NewRequest("POST", "/users/1/update-password", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorWeakPassword)

		data = `{"password":"invalid","new_password":"super_super_secure_password"}`
		r = httptest.NewRequest("POST", "/users/1/update-password", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorIncorrectPassword)

		data = `{"password":"super_secure_password","new_password":"super_super_secure_password"}`
		r = httptest.NewRequest("POST", "/users/1/update-password", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("post /users/userid/register-totp", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/users/1/register-totp")

		db := initializeTestDB(t)
		defer db.Close()

		user1 := User{
			Id:             "1",
			CreatedAt:      time.Unix(time.Now().Unix(), 0),
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("POST", "/users/2/register-totp", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		data := `{"key": "moM4ZtcDvWQQIA==", "code": "123456"}`
		r = httptest.NewRequest("POST", "/users/1/register-totp", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidData)

		data = `{"key": "j1dCsnrWOnKAfyMxShUPZ9AUwes", "code": "123456"}`
		r = httptest.NewRequest("POST", "/users/1/register-totp", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidData)

		data = `{"key": "j1dCsnrWOnKAfyMxShUPZ9AUwe$=", "code": "123456"}`
		r = httptest.NewRequest("POST", "/users/1/register-totp", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidData)

		data = `{"key": "j1dCsnrWOnKAfyMxShUPZ9AUwes=", "code": "123456"}`
		r = httptest.NewRequest("POST", "/users/1/register-totp", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorIncorrectCode)

		key := make([]byte, 20)
		_, err = rand.Read(key)
		if err != nil {
			t.Fatal(err)
		}
		totp := otp.GenerateTOTP(time.Now(), key, 30*time.Second, 6)
		data = fmt.Sprintf(`{"key":"%s", "code":"%s"}`, base64.StdEncoding.EncodeToString(key), totp)
		r = httptest.NewRequest("POST", "/users/1/register-totp", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, userTOTPCredentialJSONKeys)

		key = make([]byte, 20)
		_, err = rand.Read(key)
		if err != nil {
			t.Fatal(err)
		}
		totp = otp.GenerateTOTP(time.Now(), key, 30*time.Second, 6)
		data = fmt.Sprintf(`{"key":"%s", "code":"%s"}`, base64.StdEncoding.EncodeToString(key), totp)
		r = httptest.NewRequest("POST", "/users/1/register-totp", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, userTOTPCredentialJSONKeys)
	})

	t.Run("get /user/userid/totp-credential", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "GET", "/users/1/totp-credential")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)
		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}
		user2 := User{
			Id:             "2",
			CreatedAt:      now,
			Email:          "user2@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user2)
		if err != nil {
			t.Fatal(err)
		}

		credential1 := UserTOTPCredential{
			UserId:    user1.Id,
			CreatedAt: now,
			Key:       make([]byte, 20),
		}
		err = insertUserTOTPCredential(db, &credential1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("GET", "/users/3/totp-credential", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/users/2/totp-credential", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/users/1/totp-credential", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		var result UserTOTPCredentialJSON
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatal(err)
		}
		var expected UserTOTPCredentialJSON
		err = json.Unmarshal([]byte(credential1.EncodeToJSON()), &expected)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, result)
	})

	t.Run("delete /users/userid/totp-credential", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "DELETE", "/users/1/totp-credential")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)
		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		user2 := User{
			Id:             "2",
			CreatedAt:      now,
			Email:          "user2@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user2)
		if err != nil {
			t.Fatal(err)
		}

		credential1 := UserTOTPCredential{
			UserId:    user1.Id,
			CreatedAt: now,
			Key:       make([]byte, 20),
		}
		err = insertUserTOTPCredential(db, &credential1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("DELETE", "/users/3/totp-credential", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/users/2/totp-credential", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/users/1/totp-credential", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("post /users/userid/verify-2fa/totp", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/users/1/verify-2fa/totp")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)
		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		user2 := User{
			Id:             "2",
			CreatedAt:      now,
			Email:          "user2@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user2)
		if err != nil {
			t.Fatal(err)
		}

		key := make([]byte, 20)
		rand.Read(key)
		credential1 := UserTOTPCredential{
			UserId:    user1.Id,
			CreatedAt: now,
			Key:       key,
		}
		err = insertUserTOTPCredential(db, &credential1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("POST", "/users/3/verify-2fa/totp", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("POST", "/users/2/verify-2fa/totp", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorNotAllowed)

		data := `{"code":"123456"}`
		r = httptest.NewRequest("POST", "/users/1/verify-2fa/totp", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorIncorrectCode)

		totp := otp.GenerateTOTP(time.Now(), key, 30*time.Second, 6)
		data = fmt.Sprintf(`{"code":"%s"}`, totp)
		r = httptest.NewRequest("POST", "/users/1/verify-2fa/totp", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("post /users/userid/reset-2fa", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/users/1/reset-2fa")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)
		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("POST", "/users/2/reset-2fa", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		data := `{"recovery_code":"87654321"}`
		r = httptest.NewRequest("POST", "/users/1/reset-2fa", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorIncorrectCode)

		data = `{"recovery_code":"12345678"}`
		r = httptest.NewRequest("POST", "/users/1/reset-2fa", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, recoveryCodeJSONKeys)
	})

	t.Run("post /authenticate/password", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/authenticate/password")

		db := initializeTestDB(t)
		defer db.Close()

		user1 := User{
			Id:             "1",
			CreatedAt:      time.Unix(time.Now().Unix(), 0),
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("POST", "/authenticate/password", strings.NewReader(`{"email":"email","password":"12345678"}`))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidData)

		r = httptest.NewRequest("POST", "/authenticate/password", strings.NewReader(`{"email":"user2@example.com","password":"12345678"}`))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorUserNotExists)

		r = httptest.NewRequest("POST", "/authenticate/password", strings.NewReader(`{"email":"user1@example.com","password":"12345678"}`))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorIncorrectPassword)

		r = httptest.NewRequest("POST", "/authenticate/password", strings.NewReader(`{"email":"user1@example.com","password":"super_secure_password"}`))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, userJSONKeys)

		r = httptest.NewRequest("POST", "/authenticate/password", strings.NewReader(`{"email":"USER1@EXAMPLE.COM","password":"super_secure_password"}`))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, userJSONKeys)
	})

	t.Run("post /users/userid/email-verification-request", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/users/1/email-verification-request")

		db := initializeTestDB(t)
		defer db.Close()

		user1 := User{
			Id:             "1",
			CreatedAt:      time.Unix(time.Now().Unix(), 0),
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("POST", "/users/2/email-verification-request", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("POST", "/users/1/email-verification-request", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, userEmailVerificationRequestJSONKeys)

		r = httptest.NewRequest("POST", "/users/1/email-verification-request", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, userEmailVerificationRequestJSONKeys)
	})

	t.Run("get /users/userid/email-verification-request", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "GET", "/users/1/email-verification-request")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		user2 := User{
			Id:             "2",
			CreatedAt:      now,
			Email:          "user2@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user2)
		if err != nil {
			t.Fatal(err)
		}

		user3 := User{
			Id:             "3",
			CreatedAt:      now,
			Email:          "user3@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user3)
		if err != nil {
			t.Fatal(err)
		}

		verificationRequest1 := UserEmailVerificationRequest{
			UserId:    user1.Id,
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(10 * time.Minute),
		}
		err = insertUserEmailVerificationRequest(db, &verificationRequest1)
		if err != nil {
			t.Fatal(err)
		}

		verificationRequest2 := UserEmailVerificationRequest{
			UserId:    user2.Id,
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(-10 * time.Minute),
		}
		err = insertUserEmailVerificationRequest(db, &verificationRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("GET", "/users/4/email-verification-request", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/users/3/email-verification-request", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/users/2/email-verification-request", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/users/1/email-verification-request", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		var result UserEmailVerificationRequestJSON
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatal(err)
		}
		var expected UserEmailVerificationRequestJSON
		err = json.Unmarshal([]byte(verificationRequest1.EncodeToJSON()), &expected)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, result)
	})

	t.Run("delete /users/userid/email-verification-request", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "DELETE", "/users/1/email-verification-request")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		user2 := User{
			Id:             "2",
			CreatedAt:      now,
			Email:          "user2@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user2)
		if err != nil {
			t.Fatal(err)
		}

		user3 := User{
			Id:             "3",
			CreatedAt:      now,
			Email:          "user3@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user3)
		if err != nil {
			t.Fatal(err)
		}

		verificationRequest1 := UserEmailVerificationRequest{
			UserId:    user1.Id,
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(10 * time.Minute),
		}
		err = insertUserEmailVerificationRequest(db, &verificationRequest1)
		if err != nil {
			t.Fatal(err)
		}

		verificationRequest2 := UserEmailVerificationRequest{
			UserId:    user2.Id,
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(-10 * time.Minute),
		}
		err = insertUserEmailVerificationRequest(db, &verificationRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("DELETE", "/users/4/email-verification-request", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/users/3/email-verification-request", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/users/2/email-verification-request", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/users/1/email-verification-request", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("post /users/userid/verify-email", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/users/1/verify-email")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		user2 := User{
			Id:             "2",
			CreatedAt:      now,
			Email:          "user2@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user2)
		if err != nil {
			t.Fatal(err)
		}

		user3 := User{
			Id:             "3",
			CreatedAt:      now,
			Email:          "user3@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user3)
		if err != nil {
			t.Fatal(err)
		}

		verificationRequest1 := UserEmailVerificationRequest{
			UserId:    user1.Id,
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(10 * time.Minute),
		}
		err = insertUserEmailVerificationRequest(db, &verificationRequest1)
		if err != nil {
			t.Fatal(err)
		}

		verificationRequest2 := UserEmailVerificationRequest{
			UserId:    user2.Id,
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(-10 * time.Minute),
		}
		err = insertUserEmailVerificationRequest(db, &verificationRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("POST", "/users/4/verify-email", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("POST", "/users/3/verify-email", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorNotAllowed)

		r = httptest.NewRequest("POST", "/users/2/verify-email", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorNotAllowed)

		data := `{"code":"87654321"}`
		r = httptest.NewRequest("POST", "/users/1/verify-email", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorIncorrectCode)

		data = `{"code":"12345678"}`
		r = httptest.NewRequest("POST", "/users/1/verify-email", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("post /users/userid/email-update-requests", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/users/1/email-update-requests")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		data := `{"email":"email"}`
		r := httptest.NewRequest("POST", "/users/1/email-update-requests", strings.NewReader(data))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidData)

		data = `{"email":"user1@example.com"}`
		r = httptest.NewRequest("POST", "/users/1/email-update-requests", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorEmailAlreadyUsed)

		data = `{"email":"USER1@EXAMPLE.COM"}`
		r = httptest.NewRequest("POST", "/users/1/email-update-requests", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorEmailAlreadyUsed)

		data = `{"email":"user2@example.com"}`
		r = httptest.NewRequest("POST", "/users/1/email-update-requests", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, emailUpdateRequestJSONKeys)
	})

	t.Run("get /users/userid/email-update-requests", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "GET", "/users/1/email-update-requests")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1a@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		user2 := User{
			Id:             "2",
			CreatedAt:      now,
			Email:          "user2a@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user2)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest1 := EmailUpdateRequest{
			Id:        "1",
			UserId:    user1.Id,
			CreatedAt: now,
			Email:     "user1b@example.com",
			ExpiresAt: now.Add(10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest1)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest2 := EmailUpdateRequest{
			Id:        "2",
			UserId:    user1.Id,
			CreatedAt: now,
			Email:     "user1c@example.com",
			ExpiresAt: now.Add(-10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest2)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest3 := EmailUpdateRequest{
			Id:        "3",
			UserId:    user2.Id,
			CreatedAt: now,
			Email:     "user2b@example.com",
			ExpiresAt: now.Add(10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest3)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("GET", "/users/3/email-update-requests", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/users/1/email-update-requests", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		var result []EmailUpdateRequestJSON
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatal(err)
		}

		var expected1 EmailUpdateRequestJSON
		err = json.Unmarshal([]byte(updateRequest1.EncodeToJSON()), &expected1)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, []EmailUpdateRequestJSON{expected1}, result)
	})

	t.Run("delete /users/userid/email-update-requests", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "DELETE", "/users/1/email-update-requests")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1a@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest := EmailUpdateRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			Email:     "user1b@example.com",
			ExpiresAt: now.Add(10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("DELETE", "/users/2/email-update-requests", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/users/1/email-update-requests", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("get /email-update-requests/requestid", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "GET", "/email-update-requests/1")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1a@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest1 := EmailUpdateRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			Email:     "user1b@example.com",
			ExpiresAt: now.Add(10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest1)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest2 := EmailUpdateRequest{
			Id:        "2",
			UserId:    user.Id,
			CreatedAt: now,
			Email:     "user1c@example.com",
			ExpiresAt: now.Add(-10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("GET", "/email-update-requests/3", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/email-update-requests/2", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/email-update-requests/1", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		var result EmailUpdateRequestJSON
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatal(err)
		}
		var expected EmailUpdateRequestJSON
		err = json.Unmarshal([]byte(updateRequest1.EncodeToJSON()), &expected)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, result)
	})

	t.Run("delete /email-update-requests/requestid", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "DELETE", "/email-update-requests/1")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1a@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest1 := EmailUpdateRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			Email:     "user1b@example.com",
			ExpiresAt: now.Add(10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest1)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest2 := EmailUpdateRequest{
			Id:        "2",
			UserId:    user.Id,
			CreatedAt: now,
			Email:     "user1c@example.com",
			ExpiresAt: now.Add(-10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("DELETE", "/email-update-requests/3", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/email-update-requests/2", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/email-update-requests/1", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("post /update-email", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/update-email")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1a@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest1 := EmailUpdateRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			Email:     "user1b@example.com",
			ExpiresAt: now.Add(10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest1)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest2 := EmailUpdateRequest{
			Id:        "2",
			UserId:    user.Id,
			CreatedAt: now,
			Email:     "user1c@example.com",
			ExpiresAt: now.Add(-10 * time.Minute),
			Code:      "12345678",
		}
		err = insertEmailUpdateRequest(db, context.Background(), &updateRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		data := `{"request_id":"3","code":"123445678"}`
		r := httptest.NewRequest("POST", "/update-email", strings.NewReader(data))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidRequest)

		data = `{"request_id":"2","code":"123445678"}`
		r = httptest.NewRequest("POST", "/update-email", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidRequest)

		data = `{"request_id":"1","code":"87654321"}`
		r = httptest.NewRequest("POST", "/update-email", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorIncorrectCode)

		data = `{"request_id":"1","code":"12345678"}`
		r = httptest.NewRequest("POST", "/update-email", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		var result EmailJSON
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatal(err)
		}
		var expected EmailJSON
		err = json.Unmarshal([]byte(encodeEmailToJSON(updateRequest1.Email)), &expected)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, result)
	})

	t.Run("post /password-reset-requests", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/password-reset-requests")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		data := `{"email":"email"}`
		r := httptest.NewRequest("POST", "/password-reset-requests", strings.NewReader(data))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidData)

		data = `{"email":"user2@example.com"}`
		r = httptest.NewRequest("POST", "/password-reset-requests", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorUserNotExists)

		data = `{"email":"user1@example.com"}`
		r = httptest.NewRequest("POST", "/password-reset-requests", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, passwordResetRequestWithCodeJSONKeys)

		data = `{"email":"USER1@EXAMPLE.COM"}`
		r = httptest.NewRequest("POST", "/password-reset-requests", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertJSONResponse(t, res, passwordResetRequestWithCodeJSONKeys)
	})

	t.Run("get /password-reset-requests/requestid", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/password-reset-requests/1")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest1 := PasswordResetRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest1)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest2 := PasswordResetRequest{
			Id:        "2",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(-10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("GET", "/password-reset-requests/3", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/password-reset-requests/2", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/password-reset-requests/1", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		var result PasswordResetRequestJSON
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatal(err)
		}
		var expected PasswordResetRequestJSON
		err = json.Unmarshal([]byte(resetRequest1.EncodeToJSON()), &expected)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, result)
	})

	t.Run("delete /password-reset-requests/requestid", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "DELETE", "/password-reset-requests/1")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest1 := PasswordResetRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest1)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest2 := PasswordResetRequest{
			Id:        "2",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(-10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("DELETE", "/password-reset-requests/3", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/password-reset-requests/2", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/password-reset-requests/1", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("get /users/userid/password-reset-requests", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "GET", "/users/1/password-reset-requests")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user1 := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1a@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		user2 := User{
			Id:             "2",
			CreatedAt:      now,
			Email:          "user2a@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user2)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest1 := PasswordResetRequest{
			Id:        "1",
			UserId:    user1.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest1)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest2 := PasswordResetRequest{
			Id:        "2",
			UserId:    user1.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(-10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest2)
		if err != nil {
			t.Fatal(err)
		}

		updateRequest3 := PasswordResetRequest{
			Id:        "3",
			UserId:    user2.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &updateRequest3)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("GET", "/users/3/password-reset-requests", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("GET", "/users/1/password-reset-requests", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		var result []PasswordResetRequestJSON
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatal(err)
		}

		var expected1 PasswordResetRequestJSON
		err = json.Unmarshal([]byte(resetRequest1.EncodeToJSON()), &expected1)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, []PasswordResetRequestJSON{expected1}, result)
	})

	t.Run("delete /users/userid/password-reset-requests", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "DELETE", "/users/1/password-reset-requests")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest := PasswordResetRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		r := httptest.NewRequest("DELETE", "/users/2/password-reset-requests", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		r = httptest.NewRequest("DELETE", "/users/1/password-reset-requests", nil)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("post /password-reset-requests/requestid/verify-email", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/password-reset-requests/1/verify-email")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest1 := PasswordResetRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "$argon2id$v=19$m=19456,t=2,p=1$IQbeg/QvpmoSTQNW57r+6A$2ZzKyEAX9kU5+2S/Xv8zwjuNo9D+94a90Q1GujdgtQQ",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest1)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest2 := PasswordResetRequest{
			Id:        "2",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(-10 * time.Minute),
			CodeHash:  "$argon2id$v=19$m=19456,t=2,p=1$IQbeg/QvpmoSTQNW57r+6A$2ZzKyEAX9kU5+2S/Xv8zwjuNo9D+94a90Q1GujdgtQQ",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		data := `{"code":"123445678"}`
		r := httptest.NewRequest("POST", "/password-reset-requests/3/verify-email", strings.NewReader(data))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		data = `{"code":"123445678"}`
		r = httptest.NewRequest("POST", "/password-reset-requests/2/verify-email", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 404, "NOT_FOUND")

		data = `{"code":"87654321"}`
		r = httptest.NewRequest("POST", "/password-reset-requests/1/verify-email", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorIncorrectCode)

		data = `{"code":"12345678"}`
		r = httptest.NewRequest("POST", "/password-reset-requests/1/verify-email", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})

	t.Run("/reset-password", func(t *testing.T) {
		t.Parallel()

		testAuthentication(t, "POST", "/reset-password")

		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			CreatedAt:      now,
			Email:          "user1@example.com",
			PasswordHash:   "HASH",
			RecoveryCode:   "12345678",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest1 := PasswordResetRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest1)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest2 := PasswordResetRequest{
			Id:        "2",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(-10 * time.Minute),
			CodeHash:  "HASH",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest2)
		if err != nil {
			t.Fatal(err)
		}

		env := createEnvironment(db, nil)
		app := CreateApp(env)

		data := `{"request_id":"3","password":"123445678"}`
		r := httptest.NewRequest("POST", "/reset-password", strings.NewReader(data))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res := w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidRequest)

		data = `{"request_id":"2","password":"123445678"}`
		r = httptest.NewRequest("POST", "/reset-password", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorInvalidRequest)

		data = `{"request_id":"1","password":"123445678"}`
		r = httptest.NewRequest("POST", "/reset-password", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assertErrorResponse(t, res, 400, ExpectedErrorWeakPassword)

		data = `{"request_id":"1","password":"super_secure_password"}`
		r = httptest.NewRequest("POST", "/reset-password", strings.NewReader(data))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, r)
		res = w.Result()
		assert.Equal(t, 204, res.StatusCode)
	})
}

func TestApp(t *testing.T) {
	t.Parallel()

	db := initializeTestDB(t)
	defer db.Close()
	env := createEnvironment(db, nil)
	app := CreateApp(env)

	// Create user
	r := httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"user1a@example.com","password":"super_secure_password"}`))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res := w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /users status code")
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var user UserJSON
	err = json.Unmarshal(body, &user)
	if err != nil {
		t.Fatal(err)
	}

	// Authenticate user
	r = httptest.NewRequest("POST", "/authenticate/password", strings.NewReader(`{"email":"user1a@example.com","password":"super_secure_password"}`))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /authenticate/password status code")

	// Create email verification request
	r = httptest.NewRequest("POST", fmt.Sprintf("/users/%s/email-verification-request", user.Id), nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /users/[user_id]/email-verification-requests status code")
	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var emailVerificationRequest EmailUpdateRequestJSON
	err = json.Unmarshal(body, &emailVerificationRequest)
	if err != nil {
		t.Fatal(err)
	}

	// Verify user email
	url := fmt.Sprintf("/users/%s/verify-email", user.Id)
	data := fmt.Sprintf(`{"code":"%s"}`, emailVerificationRequest.Code)
	r = httptest.NewRequest("POST", url, strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 204, res.StatusCode, "POST /users/[user_id]/verify-email status code")

	// Update password
	r = httptest.NewRequest("POST", fmt.Sprintf("/users/%s/update-password", user.Id), strings.NewReader(`{"password":"super_secure_password","new_password":"super_secure_password_updated"}`))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 204, res.StatusCode, "POST /users/[user_id]/update-password status code")

	// Authenticate with updated password
	r = httptest.NewRequest("POST", "/authenticate/password", strings.NewReader(`{"email":"user1a@example.com","password":"super_secure_password_updated"}`))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /authenticate/password status code")

	// Create password reset request
	data = `{"email":"user1a@example.com"}`
	r = httptest.NewRequest("POST", "/password-reset-requests", strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /password-reset-requests status code")
	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var passwordResetRequestWithCode PasswordResetRequestWithCodeJSON
	err = json.Unmarshal(body, &passwordResetRequestWithCode)
	if err != nil {
		t.Fatal(err)
	}

	// Create email update request
	url = fmt.Sprintf("/users/%s/email-update-requests", user.Id)
	data = `{"email":"user1b@example.com"}`
	r = httptest.NewRequest("POST", url, strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /users/[user_id]/email-update-requests status code")
	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var emailUpdateRequest EmailUpdateRequestJSON
	err = json.Unmarshal(body, &emailUpdateRequest)
	if err != nil {
		t.Fatal(err)
	}

	// Update email
	data = fmt.Sprintf(`{"request_id":"%s","code":"%s"}`, emailUpdateRequest.Id, emailUpdateRequest.Code)
	r = httptest.NewRequest("POST", "/update-email", strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /users/[user_id]/update-email status code")

	// Get password reset request created before email update
	url = fmt.Sprintf("/password-reset-requests/%s", passwordResetRequestWithCode.Id)
	r = httptest.NewRequest("GET", url, nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 404, res.StatusCode, "GET /password-reset-requests/requestid status code")

	// Authenticate user with updated email
	r = httptest.NewRequest("POST", "/authenticate/password", strings.NewReader(`{"email":"user1b@example.com","password":"super_secure_password_updated"}`))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /authenticate/password status code")

	// Create password reset request
	data = `{"email":"user1b@example.com"}`
	r = httptest.NewRequest("POST", "/password-reset-requests", strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /password-reset-requests status code")
	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(body, &passwordResetRequestWithCode)
	if err != nil {
		t.Fatal(err)
	}

	// Verify password reset email
	url = fmt.Sprintf("/password-reset-requests/%s/verify-email", passwordResetRequestWithCode.Id)
	data = fmt.Sprintf(`{"code":"%s"}`, passwordResetRequestWithCode.Code)
	r = httptest.NewRequest("POST", url, strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 204, res.StatusCode, "POST /password-reset-requests/[request_id]/verify-email status code")

	// Reset password
	data = fmt.Sprintf(`{"request_id":"%s","password":"super_secure_password_new"}`, passwordResetRequestWithCode.Id)
	r = httptest.NewRequest("POST", "/reset-password", strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 204, res.StatusCode, "/reset-password status code")

	// Authenticate user with new password
	r = httptest.NewRequest("POST", "/authenticate/password", strings.NewReader(`{"email":"user1b@example.com","password":"super_secure_password_new"}`))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /authenticate/password status code")

	// Register TOTP credential
	key := make([]byte, 20)
	_, err = rand.Read(key)
	if err != nil {
		t.Fatal(err)
	}
	totp := otp.GenerateTOTP(time.Now(), key, 30*time.Second, 6)
	url = fmt.Sprintf("/users/%s/register-totp", user.Id)
	data = fmt.Sprintf(`{"key":"%s","code":"%s"}`, base64.StdEncoding.EncodeToString(key), totp)
	r = httptest.NewRequest("POST", url, strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /users/[user_id]/register-totp status code")

	// Verify TOTP
	totp = otp.GenerateTOTP(time.Now(), key, 30*time.Second, 6)
	url = fmt.Sprintf("/users/%s/verify-2fa/totp", user.Id)
	data = fmt.Sprintf(`{"code":"%s"}`, totp)
	r = httptest.NewRequest("POST", url, strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 204, res.StatusCode, "POST /users/[user_id]/verify-2fa/totp status code")

	// Reset 2FA credentials
	url = fmt.Sprintf("/users/%s/reset-2fa", user.Id)
	data = fmt.Sprintf(`{"recovery_code":"%s"}`, user.RecoveryCode)
	r = httptest.NewRequest("POST", url, strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /users/[user_id]/reset-2fa status code")
	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var recoveryCodeResult RecoveryCodeJSON
	err = json.Unmarshal(body, &recoveryCodeResult)
	if err != nil {
		t.Fatal(err)
	}

	// Use regenerated recovery code
	url = fmt.Sprintf("/users/%s/reset-2fa", user.Id)
	data = fmt.Sprintf(`{"recovery_code":"%s"}`, recoveryCodeResult.RecoveryCode)
	r = httptest.NewRequest("POST", url, strings.NewReader(data))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode, "POST /users/[user_id]/reset-2fa status code")
}

func assertErrorResponse(t *testing.T, res *http.Response, expectedStatus int, expectedError string) {
	assert.Equal(t, expectedStatus, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var errorData ErrorJSON
	err = json.Unmarshal(body, &errorData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedError, errorData.Error)
}

// TODO: Get JSON keys from json tags in structs?
func assertJSONResponse(t *testing.T, res *http.Response, jsonKeys []string) {
	assert.Equal(t, 200, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var responseData map[string]any
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		t.Fatal(err)
	}
	for key := range responseData {
		assert.Contains(t, jsonKeys, key)
	}
	for _, key := range jsonKeys {
		assert.Contains(t, responseData, key)
	}
}

var userJSONKeys = []string{"id", "email", "created_at", "totp_registered", "recovery_code"}
var userTOTPCredentialJSONKeys = []string{"user_id", "created_at", "key"}
var recoveryCodeJSONKeys = []string{"recovery_code"}
var userEmailVerificationRequestJSONKeys = []string{"user_id", "created_at", "expires_at", "code"}
var emailUpdateRequestJSONKeys = []string{"id", "user_id", "created_at", "email", "expires_at", "code"}
var passwordResetRequestWithCodeJSONKeys = []string{"id", "user_id", "created_at", "expires_at", "code"}

func testAuthentication(t *testing.T, method string, url string) {
	env := createEnvironment(nil, []byte("hello"))
	app := CreateApp(env)
	r := httptest.NewRequest(method, url, nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	res := w.Result()
	assertErrorResponse(t, res, 401, "NOT_AUTHENTICATED")
}
