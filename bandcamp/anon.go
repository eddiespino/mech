package bandcamp

import (
   "encoding/json"
   "fmt"
   "github.com/89z/mech"
   "net/http"
   "regexp"
   "strconv"
   "time"
)

const (
   MobileBand = "http://bandcamp.com/api/mobile/24/band_details"
   MobileTralbum = "http://bandcamp.com/api/mobile/24/tralbum_details"
)

var Heights = map[int]int{
   100: 3,
   124: 8,
   135: 15,
   138: 12,
   150: 7,
   172: 11,
   210: 9,
   300: 4,
   350: 2,
   368: 14,
   380: 13,
   700: 5,
   1200: 10,
   1500: 1,
}

var Verbose = mech.Verbose

func ArtUrl(id, height int) string {
   hID := Heights[height]
   return fmt.Sprintf("http://f4.bcbits.com/img/a%v_%v.jpg", id, hID)
}

type Band struct {
   Artists []struct {
      ID int
      Name string
   }
   Bandcamp_URL string
   Discography []Item
}

// ID to Band. Request is anonymous.
func NewBand(id int) (*Band, error) {
   req, err := http.NewRequest("GET", MobileBand, nil)
   if err != nil {
      return nil, err
   }
   val := req.URL.Query()
   val.Set("band_id", strconv.Itoa(id))
   req.URL.RawQuery = val.Encode()
   res, err := mech.RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   ban := new(Band)
   if err := json.NewDecoder(res.Body).Decode(ban); err != nil {
      return nil, err
   }
   return ban, nil
}

type Item struct {
   Item_ID int
   Item_Type string
}

// URL to Item. Request is anonymous.
func NewItem(addr string) (*Item, error) {
   req, err := http.NewRequest("HEAD", addr, nil)
   if err != nil {
      return nil, err
   }
   if req.URL.Path == "" {
      req.URL.Path = "/music"
   }
   res, err := mech.RoundTrip(req)
   if err != nil {
      return nil, err
   }
   reg := regexp.MustCompile(`nilZ0([ait])(\d+)x`)
   for _, c := range res.Cookies() {
      if c.Name == "session" {
         // [nilZ0t2809477874x t 2809477874]
         find := reg.FindStringSubmatch(c.Value)
         if find != nil {
            id, err := strconv.Atoi(find[2])
            if err == nil {
               return &Item{
                  id, find[1],
               }, nil
            }
         }
      }
   }
   return nil, fmt.Errorf("cookies %v", res.Cookies())
}

// Item to Tralbum. Request is anonymous.
func (i Item) Tralbum() (*Tralbum, error) {
   if i.Item_Type == "" {
      return nil, fmt.Errorf("%+v", i)
   }
   req, err := http.NewRequest("GET", MobileTralbum, nil)
   if err != nil {
      return nil, err
   }
   val := req.URL.Query()
   val.Set("band_id", "1")
   val.Set("tralbum_id", strconv.Itoa(i.Item_ID))
   val.Set("tralbum_type", i.Item_Type[:1])
   req.URL.RawQuery = val.Encode()
   res, err := mech.RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   tra := new(Tralbum)
   if err := json.NewDecoder(res.Body).Decode(tra); err != nil {
      return nil, err
   }
   return tra, nil
}

// All fields available with Track and Album
type Tralbum struct {
   Art_ID int
   Release_Date int64
   Title string
   Tracks []struct {
      Streaming_URL struct {
         MP3_128 string `json:"mp3-128"`
      }
   }
   Tralbum_Artist string
}

func (t Tralbum) Unix() time.Time {
   return time.Unix(t.Release_Date, 0)
}
