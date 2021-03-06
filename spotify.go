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

// Package spotify provides utilties for interfacing
// with Spotify's Web API.
package spotify

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	// DateLayout can be used with time.Parse to create time.Time values
	// from Spotify date strings.  For example, PrivateUser.Birthdate
	// uses this format.
	DateLayout = "2006-01-02"
	// TimestampLayout can be used with time.Parse to create time.Time
	// values from SpotifyTimestamp strings.  It is an ISO 8601 UTC timestamp
	// with a zero offset.  For example, PlaylistTrack's AddedAt field uses
	// this format.
	TimestampLayout = "2006-01-02T15:04:05Z"
)

var (
	baseAddress = "https://api.spotify.com/v1/"

	// DefaultClient is the default client that is used by the wrapper functions
	// that don't require authorization.  If you need to authenticate, create
	// your own client with `Authenticator.NewClient`.
	DefaultClient = &Client{
		http: new(http.Client),
	}
)

// URI identifies an artist, album, track, or category.  For example,
// spotify:track:6rqhFgbbKwnb9MLmUQDhG6
type URI string

// ID is a base-62 identifier for an artist, track, album, etc.
// It can be found at the end of a spotify.URI.
type ID string

func (id *ID) String() string {
	return string(*id)
}

// Followers contains information about the number of people following a
// particular artist or playlist.
type Followers struct {
	// The total number of followers.
	Count uint `json:"total"`
	// A link to the Web API endpoint providing full details of the followers,
	// or the empty string if this data is not available.
	Endpoint string `json:"href"`
}

// Image identifies an image associated with an item.
type Image struct {
	// The image height, in pixels.
	Height int `json:"height"`
	// The image width, in pixels.
	Width int `json:"width"`
	// The source URL of the image.
	URL string `json:"url"`
}

// Download downloads the image and writes its data to the specified io.Writer.
func (i Image) Download(dst io.Writer) error {
	resp, err := http.Get(i.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// TODO: get Content-Type from header?
	if resp.StatusCode != http.StatusOK {
		return errors.New("Couldn't download image - HTTP" + strconv.Itoa(resp.StatusCode))
	}
	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// Error represents an error returned by the Spotify Web API.
type Error struct {
	// A short description of the error.
	Message string `json:"message"`
	// The HTTP status code.
	Status int `json:"status"`
}

func (e Error) Error() string {
	return e.Message
}

// decodeError decodes an Error from an io.Reader.
func decodeError(r io.Reader) error {
	var e struct {
		E Error `json:"error"`
	}
	err := json.NewDecoder(r).Decode(&e)
	if err != nil {
		return errors.New("spotify: couldn't decode error")
	}
	return e.E
}

// ExternalID contains information that identifies an item.
type ExternalID struct {
	// The identifier type, for example:
	//   "isrc" - International Standard Recording Code
	//   "ean"  - International Article Number
	//   "upc"  - Universal Product Code
	Key string `json:"{key}"`
	// An external identifier for the object.
	Value string `json:"{value}"`
}

// ExternalURL indicates an external, public URL for an item.
type ExternalURL struct {
	// The type of the URL, for example:
	//    "spotify" - The Spotify URL for the object.
	Key string `json:"{key}"`
	// An external, public URL to the object.
	Value string `json:"{value}"`
}

// Client is a client for working with the Spotify Web API.
// To create an authenticated client, use the
// `Authenticator.NewClient` method.  If you don't need to
// authenticate, you can use `DefaultClient`.
type Client struct {
	http *http.Client
}

// Options contains optional parameters that can be provided
// to various API calls.  Only the non-nil fields are used
// in queries.
type Options struct {
	// Country is an ISO 3166-1 alpha-2 country code.  Provide
	// this parameter if you want the list of returned items to
	// be relevant to a particular country.  If omitted, the
	// results will be relevant to all countries.
	Country *string
	// Limit is the maximum number of items to return.
	Limit *int
	// Offset is the index of the first item to return.  Use it
	// with Limit to get the next set of items.
	Offset *int
}

// NewReleasesOpt is like NewReleases, but it accepts optional parameters
// for filtering the results.
func (c *Client) NewReleasesOpt(opt *Options) (albums *SimpleAlbumPage, err error) {
	spotifyURL := baseAddress + "browse/new-releases"
	if opt != nil {
		v := url.Values{}
		if opt.Country != nil {
			v.Set("country", *opt.Country)
		}
		if opt.Limit != nil {
			v.Set("limit", strconv.Itoa(*opt.Limit))
		}
		if opt.Offset != nil {
			v.Set("offset", strconv.Itoa(*opt.Offset))
		}
		if params := v.Encode(); params != "" {
			spotifyURL += "?" + params
		}
	}
	resp, err := c.http.Get(spotifyURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp.Body)
	}
	var result SimpleAlbumPage
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// NewReleases gets a list of new album releases featured in Spotify.
// This call requires bearer authorization.
func (c *Client) NewReleases() (albums *SimpleAlbumPage, err error) {
	return c.NewReleasesOpt(nil)
}
