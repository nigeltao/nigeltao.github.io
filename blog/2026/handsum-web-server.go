// Copyright 2026 Nigel Tao.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build ignore

package main

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"log"
	"net/http"

	"github.com/google/wuffs/lib/handsum"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if (len(r.URL.Path) != 73) || (r.URL.Path[:9] != "/handsum/") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		srcBytes, err := base64.URLEncoding.DecodeString(r.URL.Path[9:])
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		srcImage, err := handsum.Decode(bytes.NewReader(srcBytes))
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, srcImage)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
