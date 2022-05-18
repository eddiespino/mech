package paramount

import (
   "bytes"
   "crypto/aes"
   "crypto/cipher"
   "encoding/base64"
   "encoding/hex"
   "encoding/json"
   "encoding/xml"
   "errors"
   "github.com/89z/format"
   "github.com/89z/mech/research/widevine"
   "net/http"
   "net/url"
   "os"
   "strings"
)

const (
   aes_key = "302a6a0d70a7e9b967f91d39fef3e387816e3095925ae4537bce96063311f9c5"
   tv_secret = "6c70b33080758409"
)

func newToken() (string, error) {
   key, err := hex.DecodeString(aes_key)
   if err != nil {
      return "", err
   }
   block, err := aes.NewCipher(key)
   if err != nil {
      return "", err
   }
   var (
      dst []byte
      iv [aes.BlockSize]byte
      src []byte
   )
   src = append(src, '|')
   src = append(src, tv_secret...)
   src = pad(src)
   cipher.NewCBCEncrypter(block, iv[:]).CryptBlocks(src, src)
   dst = append(dst, 0, aes.BlockSize)
   dst = append(dst, iv[:]...)
   dst = append(dst, src...)
   return base64.StdEncoding.EncodeToString(dst), nil
}

func pad(b []byte) []byte {
   bLen := aes.BlockSize - len(b) % aes.BlockSize
   for high := byte(bLen); bLen >= 1; bLen-- {
      b = append(b, high)
   }
   return b
}

var LogLevel format.LogLevel

type Session struct {
   URL string
   Ls_Session string
}

func NewSession() (*Session, error) {
   token, err := newToken()
   if err != nil {
      return nil, err
   }
   var buf strings.Builder
   buf.WriteString("https://www.paramountplus.com/apps-api/v3.0/androidphone")
   buf.WriteString("/irdeto-control/anonymous-session-token.json")
   req, err := http.NewRequest("GET", buf.String(), nil)
   if err != nil {
      return nil, err
   }
   req.URL.RawQuery = "at=" + url.QueryEscape(token)
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   sess := new(Session)
   if err := json.NewDecoder(res.Body).Decode(sess); err != nil {
      return nil, err
   }
   return sess, nil
}
func (m Media) kID() ([]byte, error) {
   for _, ada := range m.Period.AdaptationSet {
      for _, con := range ada.ContentProtection {
         con.Default_KID = strings.ReplaceAll(con.Default_KID, "-", "")
         return hex.DecodeString(con.Default_KID)
      }
   }
   return nil, errors.New("no init data found")
}

type Media struct {
   Period struct {
      AdaptationSet []struct {
         ContentProtection []struct {
            Default_KID string `xml:"default_KID,attr"`
         }
      }
   }
}

func NewMedia(name string) (*Media, error) {
   file, err := os.Open(name)
   if err != nil {
      return nil, err
   }
   defer file.Close()
   med := new(Media)
   if err := xml.NewDecoder(file).Decode(med); err != nil {
      return nil, err
   }
   return med, nil
}

func (m Media) Keys(contentID, bearer string) ([]widevine.KeyContainer, error) {
   privateKey, err := os.ReadFile("ignore/device_private_key")
   if err != nil {
      return nil, err
   }
   clientID, err := os.ReadFile("ignore/device_client_id_blob")
   if err != nil {
      return nil, err
   }
   kID, err := m.kID()
   if err != nil {
      return nil, err
   }
   mod, err := widevine.NewModule(privateKey, clientID, kID)
   if err != nil {
      return nil, err
   }
   signedLicense, err := mod.SignedLicenseRequest()
   if err != nil {
      return nil, err
   }
   req, err := http.NewRequest(
      "POST", "https://cbsi.live.ott.irdeto.com/widevine/getlicense",
      bytes.NewReader(signedLicense),
   )
   if err != nil {
      return nil, err
   }
   req.Header.Set("Authorization", "Bearer " + bearer)
   req.URL.RawQuery = url.Values{
      "AccountId": {"cbsi"},
      "ContentId": {contentID},
   }.Encode()
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      return nil, errors.New(res.Status)
   }
   return mod.Keys(res.Body)
}

