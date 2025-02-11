package bamboo_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	bamboo "github.com/stefan-kiss/go-bamboo"
)

var (
	testComment      = "hello world"
	resultCommentKey = "TEST-TEST-1"
)

func TestAddComment(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(addCommentStub))
	defer ts.Close()

	client := bamboo.NewSimpleClient(nil, "", "")
	client.SetURL(ts.URL)

	comment := &bamboo.Comment{
		Content:   testComment,
		ResultKey: resultCommentKey,
	}

	success, resp, err := client.Comments.AddComment(comment)
	if err != nil {
		t.Error(err)
	}

	if success == false || resp.StatusCode != 204 {
		t.Error(fmt.Sprintf("Adding comment \"%s\" was unsuccessful. Returned %s", testComment, resp.Status))
	}
}

func addCommentStub(w http.ResponseWriter, r *http.Request) {
	comment := &bamboo.Comment{}
	expectedURI := fmt.Sprintf("/rest/api/latest/result/%s/comment.json", resultCommentKey)

	json.NewDecoder(r.Body).Decode(comment)

	if comment.Content != testComment {
		http.Error(w, "comments do not match", http.StatusBadRequest)
		return
	}

	if r.RequestURI != expectedURI {
		http.Error(w, fmt.Sprintf("URI did not match expected %s", resultCommentKey), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
