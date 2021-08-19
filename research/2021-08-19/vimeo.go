package main

import (
   "bytes"
   "encoding/json"
   "fmt"
   "github.com/89z/mech/html"
   "github.com/89z/mech/js"
   "net/http"
)

const videos = "http://developer.vimeo.com/api/reference/videos"

func playground(videoID string) (*http.Request, error) {
   fmt.Println("GET", videos)
   res, err := http.Get(videos)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   body, err := newSegment(videoID).callable(res)
   if err != nil {
      return nil, err
   }
   buf := new(bytes.Buffer)
   json.NewEncoder(buf).Encode(body)
   req, err := http.NewRequest(
      "POST", "http://developer.vimeo.com/api/playground/callable", buf,
   )
   if err != nil {
      return nil, err
   }
   req.Header.Set("Content-Type", "application/json")
   for _, c := range res.Cookies() {
      switch c.Name {
      case "XSRF-TOKEN":
         req.Header.Set("X-CSRF-TOKEN", c.Value)
      case "session":
         req.AddCookie(c)
      }
   }
   return req, nil
}

type app struct {
   ID int
}

type callable struct {
   App int `json:"app"`
   Group string `json:"group"`
   OperationID string `json:"operation_id"`
   PayloadParams string `json:"payload_params"`
   Ptoken string `json:"ptoken"`
   QueryParams string `json:"query_params"`
   Segments string `json:"segments"`
}

type reference struct {
   OpenAPI struct {
      Paths map[string]struct {
         Paths map[string]struct {
            Get struct {
               Token string `json:"x-playground-token"`
            }
         }
      }
   }
}

func (r reference) token() string {
   return r.OpenAPI.Paths["_essentials"].Paths["/videos/{video_id}"].Get.Token
}

type segment struct {
   VideoID string `json:"video_id"`
}

func newSegment(videoID string) segment {
   return segment{videoID}
}

func (s segment) callable(r *http.Response) (*callable, error) {
   lex := html.NewLexer(r.Body)
   lex.NextAttr("class", "footers")
   var script []byte
   for lex.NextTag("script") {
      script = append(script, lex.Bytes()...)
   }
   val := js.NewLexer(script).Values()
   call := new(callable)
   // app
   var apps []app
   if err := json.Unmarshal(val["apps"], &apps); err != nil {
      return nil, err
   }
   call.App = apps[0].ID
   // group
   call.Group = "videos"
   // operation_id
   call.OperationID = "get_video"
   // payload_params
   call.PayloadParams = "{}"
   // ptoken
   var ref reference
   if err := json.Unmarshal(val["reference"], &ref); err != nil {
      return nil, err
   }
   call.Ptoken = ref.token()
   // query_params
   call.QueryParams = "{}"
   // segments
   seg, err := json.Marshal(s)
   if err != nil {
      return nil, err
   }
   call.Segments = string(seg)
   // return
   return call, nil
}
