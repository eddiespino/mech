package youtube

import (
   "encoding/json"
   "fmt"
   "net/http"
   "strings"
)

type Player struct {
   StreamingData struct {
      AdaptiveFormats []Format
   }
   Microformat struct {
      PlayerMicroformatRenderer struct {
         AvailableCountries []string
         PublishDate string
      }
   }
   VideoDetails struct {
      Author string
      ShortDescription string
      Title string
      ViewCount int `json:"viewCount,string"`
   }
}

func NewPlayer(id string) (Player, error) {
   body := fmt.Sprintf(`
   {
      "videoId": %q, "context": {
         "client": {"clientName": "ANDROID", "clientVersion": "15.01"}
      }
   }
   `, id)
   req, err := http.NewRequest(
      "POST", PlayerAPI, strings.NewReader(body),
   )
   if err != nil {
      return Player{}, err
   }
   val := req.URL.Query()
   val.Set("key", "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8")
   req.URL.RawQuery = val.Encode()
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return Player{}, err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      return Player{}, fmt.Errorf("status %v", res.Status)
   }
   var play Player
   json.NewDecoder(res.Body).Decode(&play)
   return play, nil
}

func (p Player) Author() string {
   return p.VideoDetails.Author
}

func (p Player) Countries() []string {
   return p.Microformat.PlayerMicroformatRenderer.AvailableCountries
}

func (p Player) Description() string {
   return p.VideoDetails.ShortDescription
}

func (p Player) Formats() []Format {
   var formats []Format
   for _, format := range p.StreamingData.AdaptiveFormats {
      if format.ContentLength > 0 {
         formats = append(formats, format)
      }
   }
   return formats
}

func (p Player) NewFormat(itag int) (Format, error) {
   for _, format := range p.Formats() {
      if format.Itag == itag {
         return format, nil
      }
   }
   return Format{}, fmt.Errorf("itag %v", itag)
}

func (p Player) PublishDate() string {
   return p.Microformat.PlayerMicroformatRenderer.PublishDate
}

func (p Player) Title() string {
   return p.VideoDetails.Title
}

func (p Player) ViewCount() int {
   return p.VideoDetails.ViewCount
}