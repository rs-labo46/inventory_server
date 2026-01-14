package responder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type okBody struct {
	OK bool `json:"ok"`
}

type okBodyDecoded struct {
	OK bool `json:"ok"`
}

func TestWriteJSON_StatusAndContentTypeAndBody(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	WriteJSON(rec, http.StatusOK, okBody{OK: true})

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}

	gotCT := rec.Header().Get("Content-Type")
	wantCT := "application/json; charset=utf-8"
	if gotCT != wantCT {
		t.Fatalf("Content-Type mismatch: got=%s want=%s", gotCT, wantCT)
	}

	var decoded okBodyDecoded
	if err := json.NewDecoder(rec.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded.OK != true {
		t.Fatalf("body mismatch: got=%v want=%v", decoded.OK, true)
	}
}

type errorBodyDecoded struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestWriteError_BodyShape(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	WriteError(rec, http.StatusBadRequest, "INVALID_REQUEST", "JSONが不正です")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}

	gotCT := rec.Header().Get("Content-Type")
	wantCT := "application/json; charset=utf-8"
	if gotCT != wantCT {
		t.Fatalf("Content-Type mismatch: got=%s want=%s", gotCT, wantCT)
	}

	var decoded errorBodyDecoded
	if err := json.NewDecoder(rec.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if decoded.Error.Code != "INVALID_REQUEST" {
		t.Fatalf("code mismatch: got=%s want=%s", decoded.Error.Code, "INVALID_REQUEST")
	}
	if decoded.Error.Message != "JSONが不正です" {
		t.Fatalf("message mismatch: got=%s want=%s", decoded.Error.Message, "JSONが不正です")
	}
}
