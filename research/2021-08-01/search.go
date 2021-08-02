package main

import (
   "fmt"
   "github.com/89z/mech/youtube"
)

func main() {
   youtube.Verbose = true
   s, err := youtube.NewSearch("oneohtrix point never Nil Admirari")
   if err != nil {
      panic(err)
   }
   for _, i := range s.Items() {
      fmt.Println(i.TvMusicVideoRenderer)
   }
}
