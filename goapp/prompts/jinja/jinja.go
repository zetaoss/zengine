// Package jinja renders prompt templates in a small dialect of Jinja2 that lives with wikitext.
//
// It supports output ({{ name }}, {{ a.b }}, {{ name[1:3] }} sliced by characters), if/elif/else/endif with
// not/and/or, for/endfor with loop.index, loop.index0, loop.first and loop.last, raw/endraw, comments, and the
// "-" whitespace markers, with trim_blocks and lstrip_blocks on. Values are inserted as they are and never
// parsed again; nothing is escaped.
//
// Wiki syntax passes through as text: {{ ... }} is output only when its inside reads as an expression of this
// dialect, and {% ... %} is a tag only when it starts with a word. So {{틀|a}}, {{{1}}}, {{#if:x}} and {{!}} stay
// as they are, while {{ title }} is a variable. An expression naming a variable the template was not given, or
// using a feature this dialect lacks (filters, tests, set, macro), is an error, so a mistake is caught when the
// template is parsed instead of reaching the model; a wiki template that reads as a bare name ({{PAGENAME}}) goes
// inside {% raw %}.
package jinja

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// Template is a parsed template.
type Template struct {
	nodes []node
	uses  []Use
	spans []Span
}

// UseKind tells how a template uses a variable.
type UseKind string

const (
	UseOutput UseKind = "output"
	UseTest   UseKind = "test"
	UseLoop   UseKind = "loop"
)

// Use is a variable a template reads: printed, tested by if, or looped over by for.
type Use struct {
	Name string
	Kind UseKind
}

// SpanKind tells what a span of template source is.
type SpanKind string

const (
	SpanOutput  SpanKind = "output"
	SpanTag     SpanKind = "tag"
	SpanComment SpanKind = "comment"
)

// Span is a byte range of the source read as template syntax rather than text, for highlighting.
type Span struct {
	Start int      `json:"start"`
	End   int      `json:"end"`
	Kind  SpanKind `json:"kind"`
}

// Error is a parse or render error at a line and column (from 1, in characters) of the source.
type Error struct {
	Line, Col int
	Msg       string
}

func (e *Error) Error() string { return fmt.Sprintf("line %d, column %d: %s", e.Line, e.Col, e.Msg) }

// Uses lists the variables the template reads, each once per kind, in order of first use.
func (t *Template) Uses() []Use { return t.uses }

// Variables lists the names of the variables the template reads, each once, in order of first use.
func (t *Template) Variables() []string {
	var names []string
	for _, u := range t.uses {
		if !slices.Contains(names, u.Name) {
			names = append(names, u.Name)
		}
	}
	return names
}

// Spans lists the parts of the source read as template syntax, in order.
func (t *Template) Spans() []Span { return t.spans }

var forItem = regexp.MustCompile(`\{%-?\s*for\s+([A-Za-z][A-Za-z0-9_]*)\s+in\b`)

type node any

type textNode struct{ text string }

type outputNode struct {
	expr expr
	pos  int
}

type ifNode struct {
	branches []branch
	orElse   []node
}

type branch struct {
	cond cond
	body []node
}

type forNode struct {
	item string
	list path
	body []node
}

type path struct {
	names []string
	pos   int
}

type expr struct {
	path       path
	slice      bool
	start, end *int
}

// cond is an or of ands of possibly negated paths.
type cond [][]term

type term struct {
	not  bool
	path path
}

// Parse reads src. vars are the names the template may read; any other name is an error.
func Parse(src string, vars []string) (*Template, error) {
	p := &parser{src: src, vars: map[string]bool{}, lexLocals: map[string]bool{"loop": true}}
	for _, v := range vars {
		p.vars[v] = true
	}
	for _, m := range forItem.FindAllStringSubmatch(src, -1) {
		p.lexLocals[m[1]] = true
	}
	if err := p.lex(); err != nil {
		return nil, err
	}
	nodes, end, err := p.parseNodes(nil)
	if err != nil {
		return nil, err
	}
	if end != nil {
		return nil, p.errAt(end.pos, "unexpected {%% %s %%}", end.word)
	}
	return &Template{nodes: nodes, uses: p.uses, spans: p.spans}, nil
}

type parser struct {
	src       string
	vars      map[string]bool
	lexLocals map[string]bool
	locals    []string
	toks      []token
	i         int
	uses      []Use
	spans     []Span
}

type tokKind int

const (
	tokText tokKind = iota
	tokOutput
	tokTag
)

type token struct {
	kind tokKind
	text string
	pos  int
	word string
	rest string
}

func (p *parser) errAt(pos int, format string, args ...any) error {
	line, col := 1, 1
	for _, r := range p.src[:pos] {
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return &Error{Line: line, Col: col, Msg: fmt.Sprintf(format, args...)}
}

// lex splits the source into text, output and tag tokens, applying the whitespace rules.
func (p *parser) lex() error {
	src := p.src
	var text strings.Builder
	textStart := 0
	trimNext := false
	afterBlock := false
	flush := func() {
		if text.Len() > 0 {
			p.toks = append(p.toks, token{kind: tokText, text: text.String(), pos: textStart})
		}
		text.Reset()
	}
	addText := func(s string, pos int) {
		if text.Len() == 0 {
			textStart = pos
		}
		if trimNext {
			s = strings.TrimLeftFunc(s, unicode.IsSpace)
			if s == "" {
				return
			}
			trimNext = false
		}
		if afterBlock {
			afterBlock = false
			if strings.HasPrefix(s, "\r\n") {
				s = s[2:]
			} else if strings.HasPrefix(s, "\n") {
				s = s[1:]
			}
		}
		text.WriteString(s)
	}
	trimPrev := func() {
		s := strings.TrimRightFunc(text.String(), unicode.IsSpace)
		text.Reset()
		text.WriteString(s)
	}
	lstrip := func(pos int) {
		n := 0
		for k := pos - 1; k >= 0 && p.src[k] != '\n'; k-- {
			if p.src[k] != ' ' && p.src[k] != '\t' {
				return
			}
			n++
		}
		s := text.String()
		k := len(s)
		for k > 0 && n > 0 && (s[k-1] == ' ' || s[k-1] == '\t') {
			k--
			n--
		}
		text.Reset()
		text.WriteString(s[:k])
	}
	block := func(start, end int, kind SpanKind, leftTrim, rightTrim bool) {
		if leftTrim {
			trimPrev()
		} else {
			lstrip(start)
		}
		p.spans = append(p.spans, Span{Start: start, End: end, Kind: kind})
		trimNext = rightTrim
		afterBlock = !rightTrim
	}
	i := 0
	for i < len(src) {
		j := strings.IndexByte(src[i:], '{')
		if j < 0 {
			addText(src[i:], i)
			break
		}
		if j > 0 {
			addText(src[i:i+j], i)
			i += j
		}
		rest := src[i:]
		switch {
		case strings.HasPrefix(rest, "{{"):
			if end, inner, lt, rt, ok := p.readOutput(rest); ok {
				if lt {
					trimPrev()
				}
				afterBlock = false
				flush()
				p.toks = append(p.toks, token{kind: tokOutput, text: inner, pos: i})
				p.spans = append(p.spans, Span{Start: i, End: i + end, Kind: SpanOutput})
				trimNext = rt
				i += end
				continue
			}
			if strings.HasPrefix(rest, "{{{") {
				// A wiki parameter ({{{1}}}, {{{name|default}}}): its braces are text, and so is its name.
				addText("{{{", i)
				i += 3
			} else {
				addText("{{", i)
				i += 2
			}
		case strings.HasPrefix(rest, "{%"):
			end, word, args, lt, rt, ok, err := p.readTag(rest, i)
			if err != nil {
				return err
			}
			if !ok {
				addText("{%", i)
				i += 2
				continue
			}
			if word == "raw" {
				rawEnd, bodyStart, bodyEnd, endLT, endRT, err := p.readRaw(i+end, i)
				if err != nil {
					return err
				}
				block(i, i+end, SpanTag, lt, rt)
				afterBlock = false
				if body := src[bodyStart:bodyEnd]; body != "" {
					if endLT {
						body = strings.TrimRightFunc(body, unicode.IsSpace)
					}
					addText(body, bodyStart)
				}
				block(bodyEnd, rawEnd, SpanTag, endLT, endRT)
				i = rawEnd
				continue
			}
			block(i, i+end, SpanTag, lt, rt)
			flush()
			p.toks = append(p.toks, token{kind: tokTag, pos: i, word: word, rest: args})
			i += end
		case strings.HasPrefix(rest, "{#"):
			k := strings.Index(rest[2:], "#}")
			if k < 0 {
				return p.errAt(i, "comment {# is not closed by #}")
			}
			end := k + 4
			lt := strings.HasPrefix(rest, "{#-")
			rt := strings.HasSuffix(rest[:end], "-#}")
			block(i, i+end, SpanComment, lt, rt)
			i += end
		default:
			addText("{", i)
			i++
		}
	}
	flush()
	return nil
}

// readOutput reads {{ expr }} at the start of s. It reports ok only when the inside reads as an expression.
func (p *parser) readOutput(s string) (end int, inner string, lt, rt, ok bool) {
	k := strings.Index(s[2:], "}}")
	if k < 0 {
		return 0, "", false, false, false
	}
	inner = s[2 : 2+k]
	end = k + 4
	if strings.HasPrefix(inner, "-") {
		lt = true
		inner = inner[1:]
	}
	if strings.HasSuffix(inner, "-") {
		rt = true
		inner = inner[:len(inner)-1]
	}
	if !p.looksLikeExpr(strings.TrimSpace(inner)) {
		return 0, "", false, false, false
	}
	return end, inner, lt, rt, true
}

// looksLikeExpr reports whether the inside of {{ }} is meant as an expression: it starts with a variable of the
// template (then the rest must parse), or it is a dotted ASCII name alone (then an unknown name is an error).
// Anything else, such as a wiki template with parameters ({{틀|a}}, {{lc:x}}), is text.
func (p *parser) looksLikeExpr(s string) bool {
	w := 0
	for w < len(s) && (isNameChar(s[w]) || s[w] == '.') {
		w++
	}
	head := s[:w]
	if head == "" || !isNameStart(head[0]) {
		return false
	}
	first, _, _ := strings.Cut(head, ".")
	if p.vars[first] || p.lexLocals[first] {
		return true
	}
	return w == len(s)
}

func isNameStart(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

func isNameChar(c byte) bool { return isNameStart(c) || c >= '0' && c <= '9' || c == '_' }

// readTag reads {% word args %} at the start of s. It reports ok only when a word follows {%.
func (p *parser) readTag(s string, pos int) (end int, word, args string, lt, rt, ok bool, err error) {
	j := 2
	if j < len(s) && s[j] == '-' {
		lt = true
		j++
	}
	for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
		j++
	}
	if j >= len(s) || !isNameStart(s[j]) {
		return 0, "", "", false, false, false, nil
	}
	k := strings.Index(s, "%}")
	if k < 0 {
		return 0, "", "", false, false, false, p.errAt(pos, "tag {%% is not closed by %%}")
	}
	inner := s[j:k]
	if strings.HasSuffix(inner, "-") {
		rt = true
		inner = inner[:len(inner)-1]
	}
	inner = strings.TrimSpace(inner)
	w := 0
	for w < len(inner) && isNameChar(inner[w]) {
		w++
	}
	word, args = inner[:w], strings.TrimSpace(inner[w:])
	switch word {
	case "if", "elif", "else", "endif", "for", "endfor", "raw", "endraw":
	default:
		return 0, "", "", false, false, false, p.errAt(pos, "{%% %s %%} is not supported; this dialect has if, for and raw", word)
	}
	return k + 2, word, args, lt, rt, true, nil
}

// readRaw finds the {% endraw %} after a {% raw %} whose tag started at tagPos and ended at from.
func (p *parser) readRaw(from, tagPos int) (end, bodyStart, bodyEnd int, lt, rt bool, err error) {
	s := p.src
	for i := from; i < len(s); {
		k := strings.Index(s[i:], "{%")
		if k < 0 {
			break
		}
		at := i + k
		e, word, _, l, r, ok, _ := p.readTag(s[at:], at)
		if ok && word == "endraw" {
			return at + e, from, at, l, r, nil
		}
		i = at + 2
	}
	return 0, 0, 0, false, false, p.errAt(tagPos, "{%% raw %%} is not closed by {%% endraw %%}")
}

func (p *parser) known(name string) bool {
	return p.vars[name] || slices.Contains(p.locals, name)
}

func (p *parser) use(name string, kind UseKind) {
	if slices.Contains(p.locals, name) {
		return
	}
	u := Use{Name: name, Kind: kind}
	if !slices.Contains(p.uses, u) {
		p.uses = append(p.uses, u)
	}
}

type endTag struct {
	word string
	args string
	pos  int
}

// parseNodes reads nodes until a tag in stops, which it returns.
func (p *parser) parseNodes(stops []string) ([]node, *endTag, error) {
	var nodes []node
	for p.i < len(p.toks) {
		t := p.toks[p.i]
		p.i++
		switch t.kind {
		case tokText:
			nodes = append(nodes, textNode{t.text})
		case tokOutput:
			e, err := p.parseExpr(t.text, t.pos)
			if err != nil {
				return nil, nil, err
			}
			nodes = append(nodes, outputNode{expr: e, pos: t.pos})
		case tokTag:
			if slices.Contains(stops, t.word) {
				return nodes, &endTag{word: t.word, args: t.rest, pos: t.pos}, nil
			}
			switch t.word {
			case "if":
				n, err := p.parseIf(t)
				if err != nil {
					return nil, nil, err
				}
				nodes = append(nodes, n)
			case "for":
				n, err := p.parseFor(t)
				if err != nil {
					return nil, nil, err
				}
				nodes = append(nodes, n)
			default:
				return nil, nil, p.errAt(t.pos, "unexpected {%% %s %%}", t.word)
			}
		}
	}
	return nodes, nil, nil
}

func (p *parser) parseIf(t token) (node, error) {
	var n ifNode
	c, err := p.parseCond(t.rest, t.pos)
	if err != nil {
		return nil, err
	}
	for {
		body, end, err := p.parseNodes([]string{"elif", "else", "endif"})
		if err != nil {
			return nil, err
		}
		if end == nil {
			return nil, p.errAt(t.pos, "{%% if %%} is not closed by {%% endif %%}")
		}
		switch end.word {
		case "elif":
			n.branches = append(n.branches, branch{cond: c, body: body})
			if c, err = p.parseCond(end.args, end.pos); err != nil {
				return nil, err
			}
			continue
		case "else":
			if end.args != "" {
				return nil, p.errAt(end.pos, "{%% else %%} takes nothing")
			}
			n.branches = append(n.branches, branch{cond: c, body: body})
			orElse, end2, err := p.parseNodes([]string{"endif"})
			if err != nil {
				return nil, err
			}
			if end2 == nil {
				return nil, p.errAt(t.pos, "{%% if %%} is not closed by {%% endif %%}")
			}
			if end2.args != "" {
				return nil, p.errAt(end2.pos, "{%% endif %%} takes nothing")
			}
			n.orElse = orElse
			return n, nil
		default:
			if end.args != "" {
				return nil, p.errAt(end.pos, "{%% endif %%} takes nothing")
			}
			n.branches = append(n.branches, branch{cond: c, body: body})
			return n, nil
		}
	}
}

func (p *parser) parseFor(t token) (node, error) {
	f := strings.Fields(t.rest)
	if len(f) != 3 || f[1] != "in" || !isName(f[0]) {
		return nil, p.errAt(t.pos, "write {%% for item in list %%}")
	}
	if f[0] == "loop" || p.known(f[0]) {
		return nil, p.errAt(t.pos, "the loop variable %q hides another variable", f[0])
	}
	list, err := p.parsePath(f[2], t.pos)
	if err != nil {
		return nil, err
	}
	p.use(list.names[0], UseLoop)
	p.locals = append(p.locals, f[0], "loop")
	body, end, err := p.parseNodes([]string{"endfor", "else"})
	p.locals = p.locals[:len(p.locals)-2]
	if err != nil {
		return nil, err
	}
	if end == nil {
		return nil, p.errAt(t.pos, "{%% for %%} is not closed by {%% endfor %%}")
	}
	if end.word == "else" {
		return nil, p.errAt(end.pos, "{%% else %%} inside {%% for %%} is not supported")
	}
	if end.args != "" {
		return nil, p.errAt(end.pos, "{%% endfor %%} takes nothing")
	}
	return forNode{item: f[0], list: list, body: body}, nil
}

func isName(s string) bool {
	if s == "" || !isNameStart(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isNameChar(s[i]) {
			return false
		}
	}
	return true
}

func (p *parser) parsePath(s string, pos int) (path, error) {
	parts := strings.Split(s, ".")
	for _, part := range parts {
		if !isName(part) {
			return path{}, p.errAt(pos, "%q is not supported; write a variable name, a.b, or name[start:end]", s)
		}
	}
	if !p.known(parts[0]) {
		return path{}, p.errAt(pos, "%q is not a variable here; to write a wiki template as it is, put it inside {%% raw %%}...{%% endraw %%}", parts[0])
	}
	if parts[0] == "loop" && slices.Contains(p.locals, "loop") {
		if len(parts) != 2 || !slices.Contains([]string{"index", "index0", "first", "last"}, parts[1]) {
			return path{}, p.errAt(pos, "%q is not supported; loop has index, index0, first and last", s)
		}
	}
	return path{names: parts, pos: pos}, nil
}

func (p *parser) parseExpr(s string, pos int) (expr, error) {
	s = strings.TrimSpace(s)
	name := s
	var e expr
	if k := strings.IndexByte(s, '['); k >= 0 && strings.HasSuffix(s, "]") {
		name = strings.TrimSpace(s[:k])
		inner := s[k+1 : len(s)-1]
		a, b, found := strings.Cut(inner, ":")
		if !found {
			return expr{}, p.errAt(pos, "%q is not supported; write name[start:end]", s)
		}
		var err error
		if e.start, err = sliceBound(a); err != nil {
			return expr{}, p.errAt(pos, "%q: %v", s, err)
		}
		if e.end, err = sliceBound(b); err != nil {
			return expr{}, p.errAt(pos, "%q: %v", s, err)
		}
		e.slice = true
	}
	path, err := p.parsePath(name, pos)
	if err != nil {
		return expr{}, err
	}
	e.path = path
	p.use(path.names[0], UseOutput)
	return e, nil
}

func sliceBound(s string) (*int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("a slice bound is a whole number")
	}
	return &n, nil
}

func (p *parser) parseCond(s string, pos int) (cond, error) {
	if strings.TrimSpace(s) == "" {
		return nil, p.errAt(pos, "{%% if %%} needs a condition")
	}
	var c cond
	for _, or := range splitWord(s, "or") {
		var ands []term
		for _, and := range splitWord(or, "and") {
			f := strings.Fields(and)
			var t term
			if len(f) == 2 && f[0] == "not" {
				t.not = true
				f = f[1:]
			}
			if len(f) != 1 {
				return nil, p.errAt(pos, "%q is not supported; a condition is names joined by and, or, not", strings.TrimSpace(s))
			}
			path, err := p.parsePath(f[0], pos)
			if err != nil {
				return nil, err
			}
			p.use(path.names[0], UseTest)
			t.path = path
			ands = append(ands, t)
		}
		c = append(c, ands)
	}
	return c, nil
}

func splitWord(s, word string) []string {
	f := strings.Fields(s)
	var parts []string
	var cur []string
	for _, w := range f {
		if w == word {
			parts = append(parts, strings.Join(cur, " "))
			cur = nil
			continue
		}
		cur = append(cur, w)
	}
	return append(parts, strings.Join(cur, " "))
}

// Render fills the template with vars. A value is a string, a bool, a []string, or a map[string]any; a list
// prints its items on lines of their own. A variable left out of vars is empty.
func (t *Template) Render(vars map[string]any) (string, error) {
	var b strings.Builder
	r := renderer{scopes: []map[string]any{vars}}
	if err := r.nodes(&b, t.nodes); err != nil {
		return "", err
	}
	return b.String(), nil
}

type renderer struct {
	scopes []map[string]any
}

func (r *renderer) lookup(p path) any {
	var v any
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if x, ok := r.scopes[i][p.names[0]]; ok {
			v = x
			break
		}
	}
	for _, name := range p.names[1:] {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[name]
	}
	return v
}

func (r *renderer) nodes(b *strings.Builder, nodes []node) error {
	for _, n := range nodes {
		switch n := n.(type) {
		case textNode:
			b.WriteString(n.text)
		case outputNode:
			v := r.lookup(n.expr.path)
			if n.expr.slice {
				v = slice(v, n.expr.start, n.expr.end)
			}
			b.WriteString(Text(v))
		case ifNode:
			done := false
			for _, br := range n.branches {
				if r.test(br.cond) {
					if err := r.nodes(b, br.body); err != nil {
						return err
					}
					done = true
					break
				}
			}
			if !done {
				if err := r.nodes(b, n.orElse); err != nil {
					return err
				}
			}
		case forNode:
			items := list(r.lookup(n.list))
			for i, item := range items {
				r.scopes = append(r.scopes, map[string]any{
					n.item: item,
					"loop": map[string]any{"index": i + 1, "index0": i, "first": i == 0, "last": i == len(items)-1},
				})
				err := r.nodes(b, n.body)
				r.scopes = r.scopes[:len(r.scopes)-1]
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *renderer) test(c cond) bool {
	for _, ands := range c {
		ok := true
		for _, t := range ands {
			if Truthy(r.lookup(t.path)) == t.not {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// Truthy reports whether v counts as true for if: a non-blank string, true, a non-zero number, a non-empty list
// or map.
func Truthy(v any) bool {
	switch v := v.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case bool:
		return v
	case int:
		return v != 0
	case []string:
		return len(v) > 0
	case []any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	}
	return false
}

// Text is how a value prints: a list prints its items on lines of their own.
func Text(v any) string {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case []string:
		return strings.Join(v, "\n")
	case []any:
		parts := make([]string, len(v))
		for i, x := range v {
			parts[i] = Text(x)
		}
		return strings.Join(parts, "\n")
	}
	return fmt.Sprint(v)
}

func list(v any) []any {
	switch v := v.(type) {
	case []string:
		items := make([]any, len(v))
		for i, s := range v {
			items[i] = s
		}
		return items
	case []any:
		return v
	}
	return nil
}

// slice cuts a string by characters or a list by items, as Python does, negative bounds counting from the end.
func slice(v any, start, end *int) any {
	switch x := v.(type) {
	case string:
		runes := []rune(x)
		a, b := bounds(len(runes), start, end)
		return string(runes[a:b])
	case []string:
		a, b := bounds(len(x), start, end)
		return x[a:b]
	case []any:
		a, b := bounds(len(x), start, end)
		return x[a:b]
	}
	return v
}

func bounds(n int, start, end *int) (int, int) {
	at := func(p *int, def int) int {
		if p == nil {
			return def
		}
		i := *p
		if i < 0 {
			i += n
		}
		return max(0, min(n, i))
	}
	a, b := at(start, 0), at(end, n)
	if b < a {
		b = a
	}
	return a, b
}
