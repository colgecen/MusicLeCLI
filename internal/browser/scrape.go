package browser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Track is a single music entry extracted from an open browser tab.
type Track struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	VideoID  string `json:"videoId,omitempty"`  // YouTube Music only
	Duration string `json:"duration,omitempty"` // mm:ss when available
	Source   string `json:"source,omitempty"`   // query used for download (Spotify)
}

// Playlist is a collection extracted from a browser tab.
type Playlist struct {
	Name       string  `json:"name"`
	URL        string  `json:"url"`
	Cover      string  `json:"cover,omitempty"` // cover image URL when available
	TrackCount int     `json:"trackCount"`
	Tracks     []Track `json:"tracks"`
}

// Platform identifies which music service a tab belongs to.
type Platform string

const (
	PlatformYouTube Platform = "YouTube"
	PlatformSpotify Platform = "Spotify"
)

// MaxTracks is the hard cap applied to every scraped playlist.
const MaxTracks = 100

// Connect scans the running browser for a music tab of the given platform,
// then extracts its playlists (YouTube Music) or the open playlist (Spotify).
// It returns the extracted playlists and the raw track list for the first one.
func Connect(platform Platform) ([]Playlist, error) {
	endpoint, err := FindDevToolsEndpoint()
	if err != nil {
		return nil, err
	}
	client, err := Dial(endpoint)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	targets, err := client.GetTargets()
	if err != nil {
		return nil, err
	}

	var match *Target
	var matchedURL string
	for i := range targets {
		t := &targets[i]
		if t.Type != "page" {
			continue
		}
		if platform == PlatformYouTube && strings.Contains(t.URL, "music.youtube.com") {
			match = t
			matchedURL = t.URL
			break
		}
		if platform == PlatformSpotify && strings.Contains(t.URL, "open.spotify.com") {
			match = t
			matchedURL = t.URL
			break
		}
	}
	if match == nil {
		return nil, fmt.Errorf("%s tab is not open in the browser", platform)
	}

	session, err := client.Attach(match.TargetID)
	if err != nil {
		return nil, err
	}

	switch platform {
	case PlatformYouTube:
		return scrapeYouTubeMusic(client, session, matchedURL)
	case PlatformSpotify:
		return scrapeSpotify(client, session, matchedURL)
	default:
		return nil, fmt.Errorf("bilinmeyen platform: %s", platform)
	}
}

// ConnectFirst is a convenience wrapper returning the first playlist's tracks.
func ConnectFirst(platform Platform) (*Playlist, error) {
	pls, err := Connect(platform)
	if err != nil {
		return nil, err
	}
	if len(pls) == 0 {
		return nil, fmt.Errorf("no playlist found for %s", platform)
	}
	return &pls[0], nil
}

// capTracks truncates a track list to MaxTracks.
func capTracks(t []Track) []Track {
	if len(t) > MaxTracks {
		return t[:MaxTracks]
	}
	return t
}

func scrapeYouTubeMusic(client *CDPClient, session, pageURL string) ([]Playlist, error) {
	// Pull the playlist id from the page URL (?list= or &list=).
	listID := ""
	if idx := strings.Index(pageURL, "list="); idx >= 0 {
		rest := pageURL[idx+len("list="):]
		if amp := strings.Index(rest, "&"); amp >= 0 {
			rest = rest[:amp]
		}
		listID = rest
	}

	expr := ytMusicScrapeExpr(listID)
	raw, err := client.Evaluate(session, expr)
	if err != nil {
		return nil, err
	}
	var pl Playlist
	if err := json.Unmarshal(raw, &pl); err != nil {
		return nil, fmt.Errorf("yt music parse: %w", err)
	}
	pl.Tracks = capTracks(pl.Tracks)
	pl.TrackCount = len(pl.Tracks)
	return []Playlist{pl}, nil
}

// ytMusicScrapeExpr returns JS that reads ytInitialData and builds a playlist
// object with up to MaxTracks tracks. It tolerates missing data.
func ytMusicScrapeExpr(listID string) string {
	return fmt.Sprintf(`(function(){
  const MAX = %d;
  const out = { name: document.title || 'YouTube Music', url: location.href, cover: '', tracks: [] };
  const data = window.ytInitialData;
  if (!data) return out;
  // playlist title + header thumbnail
  try {
    const hdr = data.header && (data.header.musicPlaylistShelfRenderer || data.header.musicDetailHeaderRenderer);
    if (hdr) {
      if (hdr.title) out.name = hdr.title;
      const th = hdr.thumbnail && hdr.thumbnail.musicThumbnailRenderer && hdr.thumbnail.musicThumbnailRenderer.thumbnail && hdr.thumbnail.musicThumbnailRenderer.thumbnail.thumbnails;
      if (th && th.length) out.cover = th[th.length-1].url;
    }
  } catch(e) {}
  // find the playlist shelf
  let shelf;
  const walk = (o) => {
    if (shelf || !o || typeof o !== 'object') return;
    if (o.musicPlaylistShelfRenderer) { shelf = o.musicPlaylistShelfRenderer; return; }
    for (const k in o) {
      if (shelf) return;
      walk(o[k]);
    }
  };
  walk(data);
  const items = (shelf && shelf.contents) ? shelf.contents : [];
  for (const it of items) {
    if (out.tracks.length >= MAX) break;
    const r = it.musicResponsiveListItemRenderer;
    if (!r) continue;
    const vid = r.playlistItemData && r.playlistItemData.videoId;
    if (!vid) continue;
    let title = '', artist = '', dur = '', cover = '';
    const cols = r.flexColumns || [];
    const txt = (node) => {
      if (!node) return '';
      if (typeof node === 'string') return node;
      if (node.runs) return node.runs.map(r=>r.text||'').join('');
      return node.text || '';
    };
    if (cols[0] && cols[0].musicResponsiveListItemFlexColumnRenderer) {
      const fc = cols[0].musicResponsiveListItemFlexColumnRenderer;
      title = txt(fc.text);
    }
    if (cols[1] && cols[1].musicResponsiveListItemFlexColumnRenderer) {
      const fc = cols[1].musicResponsiveListItemFlexColumnRenderer;
      artist = txt(fc.text);
    }
    if (cols[2] && cols[2].musicResponsiveListItemFlexColumnRenderer) {
      const fc = cols[2].musicResponsiveListItemFlexColumnRenderer;
      dur = txt(fc.text);
    }
    const th = r.thumbnail && r.thumbnail.musicThumbnailRenderer && r.thumbnail.musicThumbnailRenderer.thumbnail && r.thumbnail.musicThumbnailRenderer.thumbnail.thumbnails;
    if (th && th.length) cover = th[th.length-1].url;
    out.tracks.push({ title: title, artist: artist, videoId: vid, duration: dur, source: 'https://music.youtube.com/watch?v=' + vid, cover: cover });
  }
  out.trackCount = out.tracks.length;
  return out;
})()`, MaxTracks)
}

func scrapeSpotify(client *CDPClient, session, pageURL string) ([]Playlist, error) {
	// Extract playlist id from the URL.
	id := ""
	if idx := strings.LastIndex(pageURL, "/playlist/"); idx >= 0 {
		rest := pageURL[idx+len("/playlist/"):]
		if q := strings.Index(rest, "?"); q >= 0 {
			rest = rest[:q]
		}
		id = rest
	}
	if id == "" {
		return nil, fmt.Errorf("spotify playlist id not found: %s", pageURL)
	}

	expr := spotifyScrapeExpr(id)
	raw, err := client.Evaluate(session, expr)
	if err != nil {
		return nil, err
	}
	var pl Playlist
	if err := json.Unmarshal(raw, &pl); err != nil {
		return nil, fmt.Errorf("spotify parse: %w", err)
	}
	if pl.Name == "" {
		pl.Name = "Spotify Playlist"
	}
	pl.Tracks = capTracks(pl.Tracks)
	pl.TrackCount = len(pl.Tracks)
	return []Playlist{pl}, nil
}

// spotifyScrapeExpr returns JS that grabs the web player access token from
// localStorage and calls the Spotify Web API for up to MaxTracks tracks, with a
// DOM-scrape fallback when the token or API call fails.
func spotifyScrapeExpr(playlistID string) string {
	return fmt.Sprintf(`(async function(){
  const MAX = %d;
  const out = { name: document.title || 'Spotify', url: location.href, cover: '', tracks: [] };
  const sleep = ms => new Promise(r=>setTimeout(r,ms));
  let token = '';
  try {
    for (let i=0;i<localStorage.length;i++){
      const k = localStorage.key(i);
      const v = localStorage.getItem(k) || '';
      if (k.indexOf('spotify')>=0 && v.indexOf('"accessToken"')>=0) {
        const m = v.match(/"accessToken":"([^"]+)"/);
        if (m) { token = m[1]; break; }
      }
    }
  } catch(e){}
  if (token) {
    try {
      const res = await fetch('https://api.spotify.com/v1/playlists/'+%q+'/tracks?limit='+MAX+'&offset=0', {
        headers: { Authorization: 'Bearer '+token }
      });
      if (res.ok) {
        const j = await res.json();
        out.name = (j.name) || out.name;
        if (j.images && j.images.length) out.cover = j.images[j.images.length-1].url;
        for (const it of (j.items||[])) {
          const t = it.track;
          if (!t) continue;
          const artists = (t.artists||[]).map(a=>a.name).join(', ');
          out.tracks.push({ title: t.name||'', artist: artists, source: (artists?artists+' - ':'')+(t.name||'') });
        }
      }
    } catch(e){}
  }
  if (out.tracks.length === 0) {
    // DOM fallback: scrape visible track rows.
    const rows = document.querySelectorAll('[data-testid="track-row"]');
    rows.forEach((row, i) => {
      if (i >= MAX) return;
      const titleEl = row.querySelector('[data-testid="track-name"]') || row.querySelector('a[href*="/track/"]');
      const artistEl = row.querySelector('[data-testid="track-artist"]') || row.querySelector('span a[href*="/artist/"]');
      const title = titleEl ? titleEl.textContent.trim() : '';
      const artist = artistEl ? artistEl.textContent.trim() : '';
      if (title) out.tracks.push({ title: title, artist: artist, source: (artist?artist+' - ':'')+title });
    });
  }
  out.trackCount = out.tracks.length;
  return out;
})()`, MaxTracks, playlistID)
}
