// Copyright 2014, 2015 Zac Bergquist
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spotify

import (
	"fmt"
	"net/http"
	"testing"
)

const userResponse = `
{
  "display_name" : "Ronald Pompa",
  "external_urls" : {
    "spotify" : "https://open.spotify.com/user/wizzler"
    },
    "followers" : {
      "href" : null,
      "total" : 3829
    },
    "href" : "https://api.spotify.com/v1/users/wizzler",
    "id" : "wizzler",
    "images" : [ {
      "height" : null,
      "url" : "http://profile-images.scdn.co/images/userprofile/default/9d51820e73667ea5f1e97ea601cf0593b558050e",
      "width" : null
    } ],
    "type" : "user",
    "uri" : "spotify:user:wizzler"
}`

func TestUserProfile(t *testing.T) {
	client := testClientString(http.StatusOK, userResponse)
	user, err := client.GetUsersPublicProfile("wizzler")
	if err != nil {
		t.Error(err)
		return
	}
	if user.ID != "wizzler" {
		t.Error("Expected user wizzler, got ", user.ID)
	}
	if f := user.Followers.Count; f != 3829 {
		t.Errorf("Expected 3829 followers, got %d\n", f)
	}
}

func TestCurrentUser(t *testing.T) {
	json := `{
		"country" : "US",
		"display_name" : null,
		"email" : "username@domain.com",
		"external_urls" : {
			"spotify" : "https://open.spotify.com/user/username"
		},
		"followers" : {
			"href" : null,
			"total" : 0
		},
		"href" : "https://api.spotify.com/v1/users/userame",
		"id" : "username",
		"images" : [ ],
		"product" : "premium",
		"type" : "user",
		"uri" : "spotify:user:username",
		"birthdate" : "1985-05-01"
	}`
	client := testClientString(http.StatusOK, json)

	me, err := client.CurrentUser()
	if err != nil {
		t.Error(err)
		return
	}
	if me.Country != CountryUSA ||
		me.Email != "username@domain.com" ||
		me.Product != "premium" {
		t.Error("Received incorrect response")
	}
	if me.Birthdate != "1985-05-01" {
		t.Errorf("Expected '1985-05-01', got '%s'\n", me.Birthdate)
	}
}

func TestFollowUsersMissingScope(t *testing.T) {
	json := `{
		"error": {
			"status": 403,
			"message": "Insufficient client scope"
		}
	}`
	client := testClientString(http.StatusForbidden, json)
	addDummyAuth(client)

	err := client.Follow(ID("exampleuser01"))
	if serr, ok := err.(Error); !ok {
		t.Error("Expected insufficient client scope error")
	} else {
		if serr.Status != http.StatusForbidden {
			t.Error("Expected HTTP 403")
		}
	}
}

func TestFollowUsersInvalidToken(t *testing.T) {
	json := `{
		"error": {
			"status": 401,
			"message": "Invalid access token"
		}
	}`
	client := testClientString(http.StatusUnauthorized, json)
	addDummyAuth(client)

	err := client.Follow(ID("dummyID"))
	if serr, ok := err.(Error); !ok {
		t.Error("Expected invalid token error")
	} else {
		if serr.Status != http.StatusUnauthorized {
			t.Error("Expected HTTP 401")
		}
	}
}

func TestUserFollows(t *testing.T) {
	json := "[ false, true ]"
	client := testClientString(http.StatusOK, json)
	addDummyAuth(client)
	follows, err := client.CurrentUserFollows("artist", ID("74ASZWbe4lXaubB36ztrGX"), ID("08td7MxkoHQkXnWAYD8d6Q"))
	if err != nil {
		t.Error(err)
		return
	}
	if len(follows) != 2 || follows[0] || !follows[1] {
		t.Error("Incorrect result", follows)
	}
}

func TestCurrentUsersTracks(t *testing.T) {
	client := testClientFile(http.StatusOK, "test_data/current_users_tracks.txt")
	addDummyAuth(client)
	tracks, err := client.CurrentUsersTracks()
	if err != nil {
		t.Error(err)
		return
	}
	if tracks.Limit != 20 {
		t.Errorf("Expected limit 20, got %d\n", tracks.Limit)
	}
	if tracks.Endpoint != "https://api.spotify.com/v1/me/tracks?offset=0&limit=20" {
		t.Error("Endpoint incorrect")
	}
	if tracks.Total != 3 {
		t.Errorf("Expect 3 results, got %d\n", tracks.Total)
		return
	}
	if len(tracks.Tracks) != tracks.Total {
		t.Error("Didn't get expected number of results")
		return
	}
	expected := "You & I (Nobody In The World)"
	if tracks.Tracks[0].Name != expected {
		t.Errorf("Expected '%s', got '%s'\n", expected, tracks.Tracks[0].Name)
		fmt.Printf("\n%#v\n", tracks.Tracks[0])
	}
}

func TestUsersFollowedArtists(t *testing.T) {
	json := `
{
  "artists" : {
    "items" : [ {
      "external_urls" : {
        "spotify" : "https://open.spotify.com/artist/0I2XqVXqHScXjHhk6AYYRe"
      },
      "followers" : {
        "href" : null,
        "total" : 7753
      },
      "genres" : [ "swedish hip hop" ],
      "href" : "https://api.spotify.com/v1/artists/0I2XqVXqHScXjHhk6AYYRe",
      "id" : "0I2XqVXqHScXjHhk6AYYRe",
      "images" : [ {
        "height" : 640,
        "url" : "https://i.scdn.co/image/2c8c0cea05bf3d3c070b7498d8d0b957c4cdec20",
        "width" : 640
      }, {
        "height" : 300,
        "url" : "https://i.scdn.co/image/394302b42c4b894786943e028cdd46d7baaa29b7",
        "width" : 300
      }, {
        "height" : 64,
        "url" : "https://i.scdn.co/image/ca9df7225ade6e5dfc62e7076709ca3409a7cbbf",
        "width" : 64
      } ],
      "name" : "Afasi & Filthy",
      "popularity" : 54,
      "type" : "artist",
      "uri" : "spotify:artist:0I2XqVXqHScXjHhk6AYYRe"
   } ],
  "next" : "https://api.spotify.com/v1/users/thelinmichael/following?type=artist&after=0aV6DOiouImYTqrR5YlIqx&limit=20",
  "total" : 183,
    "cursors" : {
      "after" : "0aV6DOiouImYTqrR5YlIqx"
    },
   "limit" : 20,
   "href" : "https://api.spotify.com/v1/users/thelinmichael/following?type=artist&limit=20"
  }
}`
	client := testClientString(http.StatusOK, json)
	addDummyAuth(client)
	artists, err := client.CurrentUsersFollowedArtists()
	if err != nil {
		t.Fatal(err)
	}
	exp := 20
	if artists.Limit != exp {
		t.Errorf("Expected limit %d, got %d\n", exp, artists.Limit)
	}
	if a := artists.Cursor.After; a != "0aV6DOiouImYTqrR5YlIqx" {
		t.Error("Invalid 'after' cursor")
	}
	if l := len(artists.Artists); l != 1 {
		t.Fatalf("Expected 1 artist, got %d\n", l)
	}
	if n := artists.Artists[0].Name; n != "Afasi & Filthy" {
		t.Error("Got wrong artist name")
	}
}
