package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

type httpAccount struct {
	ID    uint
	Token string
}

func TestSocialAPIWorkflowHTTP(t *testing.T) {
	router := setupTestRouter(t)
	author := registerAndLoginHTTP(t, router, "http-author")
	viewer := registerAndLoginHTTP(t, router, "http-viewer")

	publish := performJSONRequest(t, router, http.MethodPost, "/video/publish", `{
		"title":"HTTP video #Go",
		"description":"workflow #feed",
		"play_url":"/static/uploads/videos/test/http.mp4",
		"cover_url":""
	}`, author.Token)
	assertHTTPStatus(t, publish, http.StatusCreated)
	var publishBody struct {
		Video struct {
			ID       uint   `json:"id"`
			CoverURL string `json:"cover_url"`
		} `json:"video"`
	}
	decodeHTTPBody(t, publish, &publishBody)
	if publishBody.Video.ID == 0 || publishBody.Video.CoverURL != "/static/uploads/covers/default.png" {
		t.Fatalf("unexpected publish response: %#v", publishBody)
	}
	videoID := publishBody.Video.ID

	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/video/getDetail", fmt.Sprintf(`{"id":%d}`, videoID), ""), http.StatusOK)
	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/video/listByAuthorID", fmt.Sprintf(`{"author_id":%d,"limit":2}`, author.ID), ""), http.StatusOK)
	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/feed/listLatest", `{"limit":2}`, ""), http.StatusOK)
	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/feed/listByTag", `{"tag_name":"#GO","limit":2}`, ""), http.StatusOK)

	likeResponse := performJSONRequest(t, router, http.MethodPost, "/like/like", fmt.Sprintf(`{"video_id":%d}`, videoID), viewer.Token)
	assertHTTPStatus(t, likeResponse, http.StatusOK)
	assertJSONNumber(t, likeResponse, "likes_count", 1)

	isLiked := performJSONRequest(t, router, http.MethodPost, "/like/isLiked", fmt.Sprintf(`{"video_id":%d}`, videoID), viewer.Token)
	assertHTTPStatus(t, isLiked, http.StatusOK)
	assertJSONBool(t, isLiked, "is_liked", true)
	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/like/listLikedVideos", `{"limit":2}`, viewer.Token), http.StatusOK)

	commentResponse := performJSONRequest(t, router, http.MethodPost, "/comment/publish", fmt.Sprintf(`{"video_id":%d,"content":"hello"}`, videoID), viewer.Token)
	assertHTTPStatus(t, commentResponse, http.StatusCreated)
	var commentBody struct {
		Comment struct {
			ID uint `json:"id"`
		} `json:"comment"`
		CommentsCount uint `json:"comments_count"`
	}
	decodeHTTPBody(t, commentResponse, &commentBody)
	if commentBody.Comment.ID == 0 || commentBody.CommentsCount != 1 {
		t.Fatalf("unexpected comment response: %#v", commentBody)
	}
	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/comment/listComment", fmt.Sprintf(`{"video_id":%d,"limit":2}`, videoID), ""), http.StatusOK)

	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/follow/follow", fmt.Sprintf(`{"vlogger_id":%d}`, author.ID), viewer.Token), http.StatusCreated)
	isFollowing := performJSONRequest(t, router, http.MethodPost, "/follow/isFollowing", fmt.Sprintf(`{"vlogger_id":%d}`, author.ID), viewer.Token)
	assertHTTPStatus(t, isFollowing, http.StatusOK)
	assertJSONBool(t, isFollowing, "is_following", true)
	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/follow/listFollower", fmt.Sprintf(`{"account_id":%d,"limit":2}`, author.ID), ""), http.StatusOK)
	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/follow/listFollowing", fmt.Sprintf(`{"account_id":%d,"limit":2}`, viewer.ID), ""), http.StatusOK)

	sendMessage := performJSONRequest(t, router, http.MethodPost, "/message/sendMsg", fmt.Sprintf(`{"to_id":%d,"content":"hello author"}`, author.ID), viewer.Token)
	assertHTTPStatus(t, sendMessage, http.StatusOK)
	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/message/listConversation", fmt.Sprintf(`{"to_id":%d,"limit":2}`, author.ID), viewer.Token), http.StatusOK)

	profileResponse := performJSONRequest(t, router, http.MethodPost, "/account/getProfile", fmt.Sprintf(`{"account_id":%d}`, author.ID), "")
	assertHTTPStatus(t, profileResponse, http.StatusOK)
	assertResponseHidesInternalFields(t, profileResponse.Body.String())

	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/comment/delete", fmt.Sprintf(`{"comment_id":%d}`, commentBody.Comment.ID), viewer.Token), http.StatusOK)
	assertHTTPStatus(t, performJSONRequest(t, router, http.MethodPost, "/follow/unfollow", fmt.Sprintf(`{"vlogger_id":%d}`, author.ID), viewer.Token), http.StatusOK)
	unlikeResponse := performJSONRequest(t, router, http.MethodPost, "/like/unlike", fmt.Sprintf(`{"video_id":%d}`, videoID), viewer.Token)
	assertHTTPStatus(t, unlikeResponse, http.StatusOK)
	assertJSONNumber(t, unlikeResponse, "likes_count", 0)
}

func TestProtectedRoutesRequireTokenHTTP(t *testing.T) {
	router := setupTestRouter(t)

	paths := []string{
		"/account/logout",
		"/video/publish",
		"/video/uploadVideo",
		"/video/uploadCover",
		"/like/like",
		"/like/unlike",
		"/like/isLiked",
		"/like/listLikedVideos",
		"/comment/publish",
		"/comment/delete",
		"/follow/isFollowing",
		"/follow/follow",
		"/follow/unfollow",
		"/message/sendMsg",
		"/message/listConversation",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			response := performJSONRequest(t, router, http.MethodPost, path, `{}`, "")
			assertHTTPStatus(t, response, http.StatusUnauthorized)
		})
	}
}

func TestAPIErrorStatusContractHTTP(t *testing.T) {
	router := setupTestRouter(t)
	user := registerAndLoginHTTP(t, router, "http-errors")
	receiver := registerAndLoginHTTP(t, router, "http-errors-receiver")

	tests := []struct {
		name   string
		path   string
		body   string
		token  string
		status int
	}{
		{name: "video not found", path: "/video/getDetail", body: `{"id":999999}`, status: http.StatusNotFound},
		{name: "like video not found", path: "/like/like", body: `{"video_id":999999}`, token: user.Token, status: http.StatusNotFound},
		{name: "unlike video not found", path: "/like/unlike", body: `{"video_id":999999}`, token: user.Token, status: http.StatusNotFound},
		{name: "comment video not found", path: "/comment/publish", body: `{"video_id":999999,"content":"hello"}`, token: user.Token, status: http.StatusNotFound},
		{name: "comment not found", path: "/comment/delete", body: `{"comment_id":999999}`, token: user.Token, status: http.StatusNotFound},
		{name: "follow yourself", path: "/follow/follow", body: fmt.Sprintf(`{"vlogger_id":%d}`, user.ID), token: user.Token, status: http.StatusUnprocessableEntity},
		{name: "follow account not found", path: "/follow/follow", body: `{"vlogger_id":999999}`, token: user.Token, status: http.StatusNotFound},
		{name: "followers account not found", path: "/follow/listFollower", body: `{"account_id":999999}`, status: http.StatusNotFound},
		{name: "followings account not found", path: "/follow/listFollowing", body: `{"account_id":999999}`, status: http.StatusNotFound},
		{name: "message yourself", path: "/message/sendMsg", body: fmt.Sprintf(`{"to_id":%d,"content":"hello"}`, user.ID), token: user.Token, status: http.StatusUnprocessableEntity},
		{name: "conversation with yourself", path: "/message/listConversation", body: fmt.Sprintf(`{"to_id":%d}`, user.ID), token: user.Token, status: http.StatusUnprocessableEntity},
		{name: "message account not found", path: "/message/sendMsg", body: `{"to_id":999999,"content":"hello"}`, token: user.Token, status: http.StatusNotFound},
		{name: "conversation account not found", path: "/message/listConversation", body: `{"to_id":999999}`, token: user.Token, status: http.StatusNotFound},
		{name: "profile not found", path: "/account/getProfile", body: `{"account_id":999999}`, status: http.StatusNotFound},
		{name: "publish invalid url", path: "/video/publish", body: `{"title":"bad","play_url":"https://example.com/a.mp4"}`, token: user.Token, status: http.StatusBadRequest},
		{name: "feed half cursor", path: "/feed/listLatest", body: `{"latest_time":1000,"latest_id":0}`, status: http.StatusBadRequest},
		{name: "duplicate username", path: "/account/register", body: `{"username":"http-errors","password":"secret123"}`, status: http.StatusConflict},
		{name: "invalid password", path: "/account/login", body: `{"username":"http-errors","password":"wrong-password"}`, status: http.StatusUnauthorized},
		{name: "comment too large", path: "/comment/publish", body: fmt.Sprintf(`{"video_id":1,"content":%q}`, strings.Repeat("评", 501)), token: user.Token, status: http.StatusBadRequest},
		{name: "message too large", path: "/message/sendMsg", body: fmt.Sprintf(`{"to_id":%d,"content":%q}`, receiver.ID, strings.Repeat("信", 1001)), token: user.Token, status: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performJSONRequest(t, router, http.MethodPost, tt.path, tt.body, tt.token)
			assertHTTPStatus(t, response, tt.status)
		})
	}
}

func TestCompositeCursorValidationHTTP(t *testing.T) {
	router := setupTestRouter(t)
	author := registerAndLoginHTTP(t, router, "http-cursor-author")
	viewer := registerAndLoginHTTP(t, router, "http-cursor-viewer")

	type routeCase struct {
		name  string
		path  string
		token string
		body  func(string) string
	}
	routes := []routeCase{
		{
			name: "videos by author",
			path: "/video/listByAuthorID",
			body: func(cursor string) string {
				return fmt.Sprintf(`{"author_id":%d,%s}`, author.ID, cursor)
			},
		},
		{
			name: "latest feed",
			path: "/feed/listLatest",
			body: func(cursor string) string {
				return fmt.Sprintf(`{%s}`, cursor)
			},
		},
		{
			name: "tag feed",
			path: "/feed/listByTag",
			body: func(cursor string) string {
				return fmt.Sprintf(`{"tag_name":"go",%s}`, cursor)
			},
		},
		{
			name:  "liked videos",
			path:  "/like/listLikedVideos",
			token: viewer.Token,
			body: func(cursor string) string {
				return fmt.Sprintf(`{%s}`, cursor)
			},
		},
		{
			name: "comments",
			path: "/comment/listComment",
			body: func(cursor string) string {
				return fmt.Sprintf(`{"video_id":1,%s}`, cursor)
			},
		},
		{
			name: "followers",
			path: "/follow/listFollower",
			body: func(cursor string) string {
				return fmt.Sprintf(`{"account_id":%d,%s}`, author.ID, cursor)
			},
		},
		{
			name: "followings",
			path: "/follow/listFollowing",
			body: func(cursor string) string {
				return fmt.Sprintf(`{"account_id":%d,%s}`, viewer.ID, cursor)
			},
		},
		{
			name:  "conversation",
			path:  "/message/listConversation",
			token: viewer.Token,
			body: func(cursor string) string {
				return fmt.Sprintf(`{"to_id":%d,%s}`, author.ID, cursor)
			},
		},
	}
	cursors := []struct {
		name   string
		fields string
	}{
		{name: "only latest_time", fields: `"latest_time":1000,"latest_id":0`},
		{name: "only latest_id", fields: `"latest_time":0,"latest_id":1`},
		{name: "negative latest_time", fields: `"latest_time":-1,"latest_id":0`},
	}

	for _, route := range routes {
		for _, cursor := range cursors {
			t.Run(route.name+"/"+cursor.name, func(t *testing.T) {
				response := performJSONRequest(t, router, http.MethodPost, route.path, route.body(cursor.fields), route.token)
				assertHTTPStatus(t, response, http.StatusBadRequest)
				assertJSONError(t, response.Body.Bytes(), "invalid cursor")
			})
		}
	}
}

func TestSocialMutationIdempotencyAndOwnershipHTTP(t *testing.T) {
	router := setupTestRouter(t)
	author := registerAndLoginHTTP(t, router, "http-invariant-author")
	actor := registerAndLoginHTTP(t, router, "http-invariant-actor")
	other := registerAndLoginHTTP(t, router, "http-invariant-other")

	publish := performJSONRequest(t, router, http.MethodPost, "/video/publish", `{
		"title":"HTTP invariant video",
		"play_url":"/static/uploads/videos/test/invariant.mp4"
	}`, author.Token)
	assertHTTPStatus(t, publish, http.StatusCreated)
	var publishBody struct {
		Video struct {
			ID uint `json:"id"`
		} `json:"video"`
	}
	decodeHTTPBody(t, publish, &publishBody)
	videoID := publishBody.Video.ID
	if videoID == 0 {
		t.Fatal("published video id is zero")
	}

	for i := 0; i < 2; i++ {
		response := performJSONRequest(t, router, http.MethodPost, "/like/like", fmt.Sprintf(`{"video_id":%d}`, videoID), actor.Token)
		assertHTTPStatus(t, response, http.StatusOK)
		assertJSONNumber(t, response, "likes_count", 1)
	}
	liked := performJSONRequest(t, router, http.MethodPost, "/like/listLikedVideos", `{}`, actor.Token)
	assertHTTPStatus(t, liked, http.StatusOK)
	var likedBody struct {
		Likes []struct {
			ID uint `json:"id"`
		} `json:"likes"`
	}
	decodeHTTPBody(t, liked, &likedBody)
	if len(likedBody.Likes) != 1 || likedBody.Likes[0].ID != videoID {
		t.Fatalf("liked videos = %#v, want one video %d", likedBody.Likes, videoID)
	}
	for i := 0; i < 2; i++ {
		response := performJSONRequest(t, router, http.MethodPost, "/like/unlike", fmt.Sprintf(`{"video_id":%d}`, videoID), actor.Token)
		assertHTTPStatus(t, response, http.StatusOK)
		assertJSONNumber(t, response, "likes_count", 0)
	}

	for i := 0; i < 2; i++ {
		response := performJSONRequest(t, router, http.MethodPost, "/follow/follow", fmt.Sprintf(`{"vlogger_id":%d}`, author.ID), actor.Token)
		assertHTTPStatus(t, response, http.StatusCreated)
	}
	followers := performJSONRequest(t, router, http.MethodPost, "/follow/listFollower", fmt.Sprintf(`{"account_id":%d}`, author.ID), "")
	assertHTTPStatus(t, followers, http.StatusOK)
	var followersBody struct {
		Accounts []struct {
			ID uint `json:"id"`
		} `json:"accounts"`
	}
	decodeHTTPBody(t, followers, &followersBody)
	if len(followersBody.Accounts) != 1 || followersBody.Accounts[0].ID != actor.ID {
		t.Fatalf("followers = %#v, want one account %d", followersBody.Accounts, actor.ID)
	}
	for i := 0; i < 2; i++ {
		response := performJSONRequest(t, router, http.MethodPost, "/follow/unfollow", fmt.Sprintf(`{"vlogger_id":%d}`, author.ID), actor.Token)
		assertHTTPStatus(t, response, http.StatusOK)
	}

	commentResponse := performJSONRequest(t, router, http.MethodPost, "/comment/publish", fmt.Sprintf(`{"video_id":%d,"content":"owned comment"}`, videoID), actor.Token)
	assertHTTPStatus(t, commentResponse, http.StatusCreated)
	var commentBody struct {
		Comment struct {
			ID uint `json:"id"`
		} `json:"comment"`
	}
	decodeHTTPBody(t, commentResponse, &commentBody)

	unauthorizedDelete := performJSONRequest(t, router, http.MethodPost, "/comment/delete", fmt.Sprintf(`{"comment_id":%d}`, commentBody.Comment.ID), other.Token)
	assertHTTPStatus(t, unauthorizedDelete, http.StatusNotFound)
	comments := performJSONRequest(t, router, http.MethodPost, "/comment/listComment", fmt.Sprintf(`{"video_id":%d}`, videoID), "")
	assertHTTPStatus(t, comments, http.StatusOK)
	var commentsBody struct {
		Comments []struct {
			ID uint `json:"id"`
		} `json:"comments"`
	}
	decodeHTTPBody(t, comments, &commentsBody)
	if len(commentsBody.Comments) != 1 || commentsBody.Comments[0].ID != commentBody.Comment.ID {
		t.Fatalf("comments after rejected delete = %#v", commentsBody.Comments)
	}

	ownedDelete := performJSONRequest(t, router, http.MethodPost, "/comment/delete", fmt.Sprintf(`{"comment_id":%d}`, commentBody.Comment.ID), actor.Token)
	assertHTTPStatus(t, ownedDelete, http.StatusOK)
	assertJSONNumber(t, ownedDelete, "comments_count", 0)
	secondDelete := performJSONRequest(t, router, http.MethodPost, "/comment/delete", fmt.Sprintf(`{"comment_id":%d}`, commentBody.Comment.ID), actor.Token)
	assertHTTPStatus(t, secondDelete, http.StatusNotFound)
}

func registerAndLoginHTTP(t *testing.T, handler http.Handler, username string) httpAccount {
	t.Helper()

	register := performJSONRequest(t, handler, http.MethodPost, "/account/register", fmt.Sprintf(`{"username":%q,"password":"secret123"}`, username), "")
	assertHTTPStatus(t, register, http.StatusCreated)
	var registerBody struct {
		ID uint `json:"id"`
	}
	decodeHTTPBody(t, register, &registerBody)

	login := performJSONRequest(t, handler, http.MethodPost, "/account/login", fmt.Sprintf(`{"username":%q,"password":"secret123"}`, username), "")
	assertHTTPStatus(t, login, http.StatusOK)
	var loginBody struct {
		Token   string `json:"token"`
		Account struct {
			ID uint `json:"id"`
		} `json:"account"`
	}
	decodeHTTPBody(t, login, &loginBody)
	if loginBody.Token == "" || loginBody.Account.ID != registerBody.ID {
		t.Fatalf("unexpected login response: %#v", loginBody)
	}
	return httpAccount{ID: loginBody.Account.ID, Token: loginBody.Token}
}

func assertHTTPStatus(t *testing.T, response interface {
	Result() *http.Response
}, want int) {
	t.Helper()
	result := response.Result()
	if result.StatusCode != want {
		t.Fatalf("status = %d, want %d", result.StatusCode, want)
	}
}

func decodeHTTPBody(t *testing.T, response interface{ Result() *http.Response }, target any) {
	t.Helper()
	defer response.Result().Body.Close()
	if err := json.NewDecoder(response.Result().Body).Decode(target); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

func assertJSONNumber(t *testing.T, response interface{ Result() *http.Response }, key string, want float64) {
	t.Helper()
	var body map[string]any
	decodeHTTPBody(t, response, &body)
	if got, ok := body[key].(float64); !ok || got != want {
		t.Fatalf("%s = %#v, want %v", key, body[key], want)
	}
}

func assertJSONBool(t *testing.T, response interface{ Result() *http.Response }, key string, want bool) {
	t.Helper()
	var body map[string]any
	decodeHTTPBody(t, response, &body)
	if got, ok := body[key].(bool); !ok || got != want {
		t.Fatalf("%s = %#v, want %t", key, body[key], want)
	}
}

func assertJSONError(t *testing.T, body []byte, want string) {
	t.Helper()

	var response struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode error response: %v; body=%s", err, body)
	}
	if response.Error != want {
		t.Fatalf("error = %q, want %q", response.Error, want)
	}
}

func assertResponseHidesInternalFields(t *testing.T, body string) {
	t.Helper()
	for _, field := range []string{"password_hash", "refresh_token", "deleted_at"} {
		if strings.Contains(body, field) {
			t.Fatalf("response exposed internal field %q: %s", field, body)
		}
	}
}
