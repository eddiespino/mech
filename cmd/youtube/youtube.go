package main

import (
   "fmt"
   "github.com/89z/mech"
   "github.com/89z/mech/youtube"
   "os"
)

func (v video) player() (*youtube.Player, error) {
   if v.id == "" {
      var err error
      v.id, err = youtube.VideoID(v.address)
      if err != nil {
         return nil, err
      }
   }
   if v.request == 1 {
      return youtube.AndroidEmbed.Player(v.id)
   }
   if v.request >= 2 {
      home, err := os.UserHomeDir()
      if err != nil {
         return nil, err
      }
      change, err := youtube.OpenExchange(home, "mech/youtube.json")
      if err != nil {
         return nil, err
      }
      if v.request == 2 {
         return youtube.AndroidRacy.Exchange(v.id, change)
      }
      return youtube.AndroidContent.Exchange(v.id, change)
   }
   return youtube.Android.Player(v.id)
}

type video struct {
   address string
   audio string
   height int
   id string
   info bool
   request int
}
func (v video) do() error {
   play, err := v.player()
   if err != nil {
      return err
   }
   forms := play.StreamingData.AdaptiveFormats
   if v.info {
      forms.MediaType()
      fmt.Println(play)
   } else {
      fmt.Println(play.PlayabilityStatus)
      if v.audio != "" {
         form, ok := forms.Audio(v.audio)
         if ok {
            err := download(form, play.Base())
            if err != nil {
               return err
            }
         }
      }
      if v.height >= 1 {
         form, ok := forms.Video(v.height)
         if ok {
            err := download(form, play.Base())
            if err != nil {
               return err
            }
         }
      }
   }
   return nil
}

func download(form *youtube.Format, base string) error {
   ext, err := mech.ExtensionByType(form.MimeType)
   if err != nil {
      return err
   }
   file, err := os.Create(base + ext)
   if err != nil {
      return err
   }
   defer file.Close()
   if _, err := form.WriteTo(file); err != nil {
      return err
   }
   return nil
}

func doRefresh() error {
   oauth, err := youtube.NewOAuth()
   if err != nil {
      return err
   }
   fmt.Println(oauth)
   fmt.Scanln()
   change, err := oauth.Exchange()
   if err != nil {
      return err
   }
   home, err := os.UserHomeDir()
   if err != nil {
      return err
   }
   return change.Create(home, "mech/youtube.json")
}

func doAccess() error {
   home, err := os.UserHomeDir()
   if err != nil {
      return err
   }
   change, err := youtube.OpenExchange(home, "mech/youtube.json")
   if err != nil {
      return err
   }
   if err := change.Refresh(); err != nil {
      return err
   }
   return change.Create(home, "mech/youtube.json")
}
