// Copyright 2019 Nigel Tao.
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

// +build ignore

package main

// update.go updates the blog posts listed in README.md.

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

func main() {
	if err := main1(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func main1() error {
	posts, err := findBlogPosts()
	if err != nil {
		return err
	}
	if len(posts) == 0 {
		return errors.New("no blog posts")
	}
	if err := writeReadme(posts); err != nil {
		return err
	}
	return nil
}

func writeReadme(posts []blogPost) error {
	dst := bytes.NewBuffer(nil)

	src, err := ioutil.ReadFile("README.md")
	if err != nil {
		return err
	}

	const blog = "\n\n## Blog\n\n"
	src = match(dst, src, blog)

	dst.WriteString(blog)
	for _, p := range posts {
		fmt.Fprintf(dst, "- %s [%s](%s)\n", p.date, p.title, p.filename)
	}

	const proj = "\n\n## Projects\n\n"
	src = match(nil, src, proj)

	const copr = "\n\nCopyright "
	src = match(dst, src, copr)
	fmt.Fprintf(dst, "\n\nCopyright %s-%s Nigel Tao, nigeltao@golang.org\n",
		posts[0].date[:4],
		posts[len(posts)-1].date[:4],
	)

	return ioutil.WriteFile("README.md", dst.Bytes(), 0666)
}

func match(dst *bytes.Buffer, src []byte, substring string) []byte {
	i := bytes.Index(src, []byte(substring))
	if i < 0 {
		i = len(src)
	}
	if dst != nil {
		dst.Write(src[:i])
	}
	return src[i:]
}

type blogPost struct {
	date     string
	filename string
	title    string
}

func findBlogPosts() (posts []blogPost, _ error) {
	infos0, err := ioutil.ReadDir("./blog")
	if err != nil {
		return nil, err
	}
	for _, f0 := range infos0 {
		infos1, err := ioutil.ReadDir("./blog/" + f0.Name())
		if err != nil {
			return nil, err
		}
		for _, f1 := range infos1 {
			if !strings.HasSuffix(f1.Name(), ".md") {
				continue
			}
			filename := "./blog/" + f0.Name() + "/" + f1.Name()
			post, err := load(filename)
			if err == errNotABlogPost {
				continue
			} else if err != nil {
				return nil, err
			}
			posts = append(posts, post)
		}
	}

	sort.Slice(posts, func(i int, j int) bool {
		if x, y := posts[i].date, posts[j].date; x != y {
			return x < y
		}
		return posts[i].filename < posts[j].filename
	})
	return posts, nil
}

func load(filename string) (blogPost, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return blogPost{}, err
	}
	for (len(b) > 0) && (b[len(b)-1] == '\n') {
		b = b[:len(b)-1]
	}

	if (len(b) < 2) || (b[0] != '#') || (b[1] != ' ') {
		return blogPost{}, errNotABlogPost
	}
	b = b[2:]

	title := ""
	if i := bytes.IndexByte(b, '\n'); i < 0 {
		return blogPost{}, errNotABlogPost
	} else {
		title = string(b[:i])
		b = b[i+1:]
	}

	const suffix0, suffix1 = "Published: ", "2000-00-00"
	if n := len(suffix0) + len(suffix1); len(b) <= n {
		return blogPost{}, errNotABlogPost
	} else {
		b = b[len(b)-n:]
	}
	if n := len(suffix0); string(b[:n]) != suffix0 {
		return blogPost{}, errNotABlogPost
	} else {
		b = b[n:]
	}

	return blogPost{
		date:     string(b),
		filename: filename,
		title:    title,
	}, nil
}

var errNotABlogPost = errors.New("not a blog post")
