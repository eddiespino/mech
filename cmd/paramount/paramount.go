package main

import (
   "fmt"
   "github.com/89z/format/hls"
   "github.com/89z/mech/paramount"
   "net/http"
   "os"
   "sort"
)

func doManifest(guid string, bandwidth int64, info bool) error {
   media, err := paramount.NewMedia(guid)
   if err != nil {
      return err
   }
   fmt.Println("GET", media.Body.Seq.Video.Src)
   res, err := http.Get(media.Body.Seq.Video.Src)
   if err != nil {
      return err
   }
   defer res.Body.Close()
   mas, err := hls.NewMaster(res.Request.URL, res.Body)
   if err != nil {
      return err
   }
   sort.Slice(mas.Stream, func(a, b int) bool {
      return mas.Stream[a].Bandwidth < mas.Stream[b].Bandwidth
   })
   if info {
      fmt.Println(media.Body.Seq.Video.Title)
      for _, video := range mas.Stream {
         video.URI = nil
         fmt.Println(video)
      }
   } else {
      video := mas.GetStream(func (s hls.Stream) bool {
         return s.Bandwidth >= bandwidth
      })
      if err := download(media, video); err != nil {
         return err
      }
   }
   return nil
}

func download(media *paramount.Media, video *hls.Stream) error {
   seg, err := newSegment(video.URI.String())
   if err != nil {
      return err
   }
   fmt.Println("GET", seg.Key.URI)
   res, err := http.Get(seg.Key.URI.String())
   if err != nil {
      return err
   }
   defer res.Body.Close()
   dec, err := hls.NewDecrypter(res.Body)
   if err != nil {
      return err
   }
   file, err := os.Create(media.Body.Seq.Video.Title + seg.Ext())
   if err != nil {
      return err
   }
   defer file.Close()
   video.URI = nil
   fmt.Println(video)
   for i, info := range seg.Info {
      fmt.Print(seg.Progress(i))
      res, err := http.Get(info.URI.String())
      if err != nil {
         return err
      }
      if _, err := dec.Copy(file, res.Body); err != nil {
         return err
      }
      if err := res.Body.Close(); err != nil {
         return err
      }
   }
   return nil
}

func newSegment(addr string) (*hls.Segment, error) {
   fmt.Println("GET", addr)
   res, err := http.Get(addr)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   return hls.NewSegment(res.Request.URL, res.Body)
}
