package db

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

type ContentType string

const (
	CT_NONE   ContentType = ""
	CT_DOCM   ContentType = "DOC"
	CT_TEXT   ContentType = "text"
	CT_BOLD   ContentType = "bold"
	CT_ITAL   ContentType = "ital"
	CT_LINK   ContentType = "link"
	CT_CODE   ContentType = "code"
	CT_STRIKE ContentType = "strike"
	CT_EMOTE  ContentType = "emote"
)

type MDNode struct {
	T       ContentType
	Content string

	Children []*MDNode
	Parent   *MDNode
}

func (t MDNode) MarshalJSON() ([]byte, error) {
	m := map[string]json.RawMessage{}
	var err error

	m["type"], err = json.Marshal(t.T)
	if err != nil {
		return nil, err
	}

	if t.Content != "" {
		m["content"], err = json.Marshal(t.Content)
		if err != nil {
			return nil, err
		}
	}
	if len(t.Children) != 0 {
		m["children"], err = json.Marshal(t.Children)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(m)
}

func makeRegEmotes(emotes map[string]string) *regexp.Regexp {
	regRaw := ``

	for n := range emotes {
		regRaw += n + "|"
	}

	if regRaw == `` {
		return nil
	}

	regRaw = regRaw[:len(regRaw)-1]

	reg := regexp.MustCompile(regRaw)
	return reg
}

func renderEmotes(r *MDNode, reg *regexp.Regexp, emotes map[string]string) []*MDNode {
	if reg == nil {
		return r.Children
	}

	if len(r.Children) != 0 {
		for _, n := range r.Children {
			n.Children = renderEmotes(n, reg, emotes)
		}
		return r.Children
	} else {
		c := strings.ToLower(r.Content)

		found := reg.FindAllStringIndex(c, 999999999999)

		renders := []*MDNode{}

		lastI := 0
		for _, f := range found {
			bfr := c[lastI:f[0]]
			if bfr != "" {
				renders = append(renders, &MDNode{
					T:       CT_TEXT,
					Content: bfr,
					Parent:  r.Parent,
				})
			}

			renders = append(renders, &MDNode{
				T:       CT_EMOTE,
				Content: emotes[c[f[0]:f[1]]],
				Parent:  r.Parent,
			})
		}

		renders = append(renders, &MDNode{
			T:       CT_TEXT,
			Content: c[lastI:],
			Parent:  r.Parent,
		})

		r.Content = ""

		return renders
	}
}

// returns true if it changed something
func simplifyTree(r *MDNode) bool {
	if r.T != CT_LINK && len(r.Children) == 1 && len(r.Children[0].Children) == 0 {
		r.Content = r.Children[0].Content
		r.Children = nil
		return true
	} else {
		newChildren := []*MDNode{}
		changed := false
		for _, n := range r.Children {
			for simplifyTree(n) {
				changed = true
			}

			if len(n.Children) == 0 && n.Content == "" {
				changed = true
				continue
			}
			newChildren = append(newChildren, n)
		}
		r.Children = newChildren

		return changed
	}
}

func ParseContent(inp string, baseEmotes []*twitch.Emote) []*MDNode {
	emotes := processEmotes(inp, baseEmotes)

	p := parser.NewWithExtensions(parser.Strikethrough | parser.Autolink)
	parsed := markdown.Parse([]byte(inp), p)

	curN := &MDNode{
		T: CT_DOCM,
	}

	ast.WalkFunc(parsed, func(node ast.Node, entering bool) ast.WalkStatus {
		elmT := CT_NONE
		if !entering {
			if curN.Parent != nil {
				curN = curN.Parent
			}
		}
		extraInfo := ""

		switch n := node.(type) {
		case *ast.Text:
			curN.Children = append(curN.Children, &MDNode{
				T:       CT_TEXT,
				Parent:  curN,
				Content: string(n.Literal),
			})
			return ast.GoToNext
		case *ast.Emph:
			elmT = CT_ITAL
		case *ast.Strong:
			elmT = CT_BOLD
		case *ast.Link:
			extraInfo = string(n.Destination)
			elmT = CT_LINK
		case *ast.Code:
			elmT = CT_CODE
		case *ast.Document, *ast.Paragraph:
			return ast.GoToNext
		case *ast.Del:
			elmT = CT_STRIKE
		default:
			return ast.SkipChildren
		}

		if entering {
			newN := &MDNode{
				T:        elmT,
				Children: []*MDNode{},
				Parent:   curN,
				Content:  extraInfo,
			}

			curN.Children = append(curN.Children, newN)

			curN = newN
		}

		return ast.GoToNext
	})

	curN.Children = renderEmotes(curN, makeRegEmotes(emotes), emotes)

	// don't simplify the root doc
	for _, n := range curN.Children {
		simplifyTree(n)
	}

	return curN.Children
}
