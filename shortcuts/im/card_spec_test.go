// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"encoding/json"
	"strings"
	"testing"
)

// compileTo is a test helper that compiles a YAML spec and decodes the
// resulting JSON back to a generic map so assertions can navigate the
// tree without a second JSON-mapping step in every test.
func compileTo(t *testing.T, src string) map[string]any {
	t.Helper()
	out, err := Compile([]byte(src))
	if err != nil {
		t.Fatalf("Compile() error = %v\nspec:\n%s", err, src)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("Compile produced invalid JSON: %v\noutput:\n%s", err, out)
	}
	return m
}

func mustElements(t *testing.T, card map[string]any) []any {
	t.Helper()
	body, ok := card["body"].(map[string]any)
	if !ok {
		t.Fatalf("card.body missing or wrong type: %#v", card["body"])
	}
	elems, ok := body["elements"].([]any)
	if !ok {
		t.Fatalf("card.body.elements missing or wrong type: %#v", body["elements"])
	}
	return elems
}

func TestCompile_MinimalCard(t *testing.T) {
	card := compileTo(t, `
elements:
  - md: "Hello world"
`)
	if card["schema"] != "2.0" {
		t.Errorf("schema = %v, want 2.0", card["schema"])
	}
	cfg, ok := card["config"].(map[string]any)
	if !ok {
		t.Fatalf("config missing")
	}
	// v2.0 default config is empty. wide_screen_mode and update_multi
	// are v1 fields and should not appear unless explicitly set.
	if _, has := cfg["wide_screen_mode"]; has {
		t.Errorf("wide_screen_mode should not appear by default in schema 2.0: %#v", cfg)
	}
	if _, has := cfg["update_multi"]; has {
		t.Errorf("update_multi should not appear by default in schema 2.0: %#v", cfg)
	}
	if _, has := card["header"]; has {
		t.Errorf("header should be absent when not declared")
	}
	elems := mustElements(t, card)
	if len(elems) != 1 {
		t.Fatalf("elements len = %d, want 1", len(elems))
	}
	e := elems[0].(map[string]any)
	if e["tag"] != "markdown" || e["content"] != "Hello world" {
		t.Errorf("md element = %#v", e)
	}
}

func TestCompile_HeaderTemplateAndSummary(t *testing.T) {
	card := compileTo(t, `
header:
  title: "📊 Report"
  subtitle: "5 items"
  template: blue
  icon: standard:notice-bell
summary: "5 items pending review"
elements:
  - md: "body"
`)
	header := card["header"].(map[string]any)
	if title := header["title"].(map[string]any); title["content"] != "📊 Report" || title["tag"] != "plain_text" {
		t.Errorf("header.title = %#v", title)
	}
	if header["template"] != "blue" {
		t.Errorf("header.template = %v", header["template"])
	}
	icon := header["icon"].(map[string]any)
	if icon["tag"] != "standard_icon" || icon["token"] != "notice-bell" {
		t.Errorf("header.icon = %#v", icon)
	}
	cfg := card["config"].(map[string]any)
	summary := cfg["summary"].(map[string]any)
	if summary["content"] != "5 items pending review" {
		t.Errorf("summary = %#v", summary)
	}
}

func TestCompile_HeaderTemplate_Invalid(t *testing.T) {
	_, err := Compile([]byte(`
header: { title: "X", template: rainbow }
elements: [{ md: "x" }]
`))
	if err == nil || !strings.Contains(err.Error(), "invalid header.template") {
		t.Errorf("error = %v, want invalid header.template", err)
	}
}

func TestCompile_AllSimpleElements(t *testing.T) {
	card := compileTo(t, `
elements:
  - md: "first"
  - hr
  - note: "small caption"
  - md:
      content: "with size"
      size: notation
  - image:
      key: img_x123
      alt: "diagram"
      compact: true
`)
	elems := mustElements(t, card)
	if len(elems) != 5 {
		t.Fatalf("elements len = %d, want 5", len(elems))
	}
	if elems[0].(map[string]any)["tag"] != "markdown" {
		t.Errorf("[0] = %#v", elems[0])
	}
	if elems[1].(map[string]any)["tag"] != "hr" {
		t.Errorf("[1] = %#v", elems[1])
	}
	noteEl := elems[2].(map[string]any)
	if noteEl["text_size"] != "notation" || noteEl["content"] != "small caption" {
		t.Errorf("note = %#v", noteEl)
	}
	mdSized := elems[3].(map[string]any)
	if mdSized["text_size"] != "notation" {
		t.Errorf("sized md = %#v", mdSized)
	}
	img := elems[4].(map[string]any)
	if img["tag"] != "img" || img["img_key"] != "img_x123" || img["compact_width"] != true {
		t.Errorf("image = %#v", img)
	}
}

func TestCompile_ActionsInlineVsEqual(t *testing.T) {
	card := compileTo(t, `
elements:
  - actions:
      - text: A
        value: a
        style: primary
      - text: B
        value: b
`)
	a := mustElements(t, card)[0].(map[string]any)
	if a["tag"] != "action" {
		t.Errorf("inline actions tag = %v", a["tag"])
	}
	btns := a["actions"].([]any)
	if len(btns) != 2 {
		t.Fatalf("inline button count = %d", len(btns))
	}
	first := btns[0].(map[string]any)
	if first["type"] != "primary" {
		t.Errorf("primary style not applied: %#v", first)
	}
	val := first["value"].(map[string]any)
	if val["action"] != "a" {
		t.Errorf("button value = %#v", val)
	}

	cardEq := compileTo(t, `
elements:
  - actions:
      - { text: A, value: a, style: primary }
      - { text: B, value: b, style: danger }
    layout: equal
`)
	cs := mustElements(t, cardEq)[0].(map[string]any)
	if cs["tag"] != "column_set" {
		t.Fatalf("equal layout should be column_set, got %v", cs["tag"])
	}
	if cs["flex_mode"] != "bisect" {
		t.Errorf("flex_mode for 2 equal buttons = %v, want bisect", cs["flex_mode"])
	}
	cols := cs["columns"].([]any)
	col0 := cols[0].(map[string]any)
	btn := col0["elements"].([]any)[0].(map[string]any)
	if btn["width"] != "fill" {
		t.Errorf("equal-layout button missing width=fill: %#v", btn)
	}
}

func TestCompile_ButtonURL_NoCallback(t *testing.T) {
	card := compileTo(t, `
elements:
  - actions:
      - text: Open
        url: "https://example.com"
`)
	btn := mustElements(t, card)[0].(map[string]any)["actions"].([]any)[0].(map[string]any)
	if btn["url"] != "https://example.com" {
		t.Errorf("url = %v", btn["url"])
	}
	if _, has := btn["value"]; has {
		t.Errorf("URL button should not have callback value: %#v", btn)
	}
}

func TestCompile_ButtonMissingValueAndURL(t *testing.T) {
	_, err := Compile([]byte(`
elements:
  - actions:
      - text: BrokenButton
`))
	if err == nil || !strings.Contains(err.Error(), "must set either value (callback) or url (link)") {
		t.Errorf("error = %v, want missing value/url message", err)
	}
}

func TestCompile_ListItem(t *testing.T) {
	card := compileTo(t, `
elements:
  - item:
      text: "**Group A**"
      btn:
        text: "Open"
        value: "open:a"
        style: primary
`)
	cs := mustElements(t, card)[0].(map[string]any)
	if cs["tag"] != "column_set" || cs["flex_mode"] != "none" {
		t.Fatalf("list item layout wrong: %#v", cs)
	}
	cols := cs["columns"].([]any)
	if len(cols) != 2 {
		t.Fatalf("list item should have 2 columns, got %d", len(cols))
	}
	left := cols[0].(map[string]any)
	if left["weight"].(float64) != 5 {
		t.Errorf("left column weight = %v, want 5", left["weight"])
	}
	right := cols[1].(map[string]any)
	if right["width"] != "auto" {
		t.Errorf("right column width = %v, want auto", right["width"])
	}
	btn := right["elements"].([]any)[0].(map[string]any)
	if btn["type"] != "primary" {
		t.Errorf("list item button style not applied: %#v", btn)
	}
}

func TestCompile_Select(t *testing.T) {
	card := compileTo(t, `
elements:
  - select:
      placeholder: "Pick one"
      value: "pick_action"
      initial: opt1
      options:
        - { text: "One", value: opt1 }
        - { text: "Two", value: opt2 }
`)
	a := mustElements(t, card)[0].(map[string]any)
	if a["tag"] != "action" {
		t.Fatalf("select wrapper tag = %v", a["tag"])
	}
	sel := a["actions"].([]any)[0].(map[string]any)
	if sel["tag"] != "select_static" {
		t.Errorf("select tag = %v", sel["tag"])
	}
	if sel["initial_option"] != "opt1" {
		t.Errorf("initial_option = %v", sel["initial_option"])
	}
	opts := sel["options"].([]any)
	if len(opts) != 2 {
		t.Errorf("options len = %d", len(opts))
	}
}

func TestCompile_PanelNested(t *testing.T) {
	card := compileTo(t, `
elements:
  - panel:
      title: "[IMPORTANT] AI Cert"
      border: red
      expanded: true
      elements:
        - md: "body"
        - hr
        - actions:
            - { text: Open, value: open }
`)
	p := mustElements(t, card)[0].(map[string]any)
	if p["tag"] != "collapsible_panel" {
		t.Fatalf("panel tag = %v", p["tag"])
	}
	if p["expanded"] != true {
		t.Errorf("panel expanded = %v", p["expanded"])
	}
	border := p["border"].(map[string]any)
	if border["color"] != "red" {
		t.Errorf("border color = %v", border["color"])
	}
	header := p["header"].(map[string]any)
	title := header["title"].(map[string]any)
	if title["tag"] != "markdown" || title["content"] != "[IMPORTANT] AI Cert" {
		t.Errorf("panel header title = %#v", title)
	}
	children := p["elements"].([]any)
	if len(children) != 3 {
		t.Errorf("panel children = %d, want 3", len(children))
	}
	if children[2].(map[string]any)["tag"] != "action" {
		t.Errorf("panel child[2] = %#v", children[2])
	}
}

func TestCompile_PanelInvalidBorder(t *testing.T) {
	_, err := Compile([]byte(`
elements:
  - panel:
      title: T
      border: purple
      elements: [{ md: x }]
`))
	if err == nil || !strings.Contains(err.Error(), "invalid panel.border") {
		t.Errorf("error = %v, want invalid border", err)
	}
}

func TestCompile_Columns(t *testing.T) {
	card := compileTo(t, `
elements:
  - columns:
      flex: bisect
      items:
        - weight: 5
          elements:
            - md: "Left"
        - weight: auto
          elements:
            - md: "Right"
`)
	cs := mustElements(t, card)[0].(map[string]any)
	if cs["tag"] != "column_set" || cs["flex_mode"] != "bisect" {
		t.Fatalf("columns = %#v", cs)
	}
	cols := cs["columns"].([]any)
	// JSON round-trip turns ints into float64.
	if cols[0].(map[string]any)["weight"].(float64) != 5 {
		t.Errorf("col[0] weight = %v", cols[0].(map[string]any)["weight"])
	}
	if cols[1].(map[string]any)["width"] != "auto" {
		t.Errorf("col[1] width = %v", cols[1].(map[string]any)["width"])
	}
}

func TestCompile_RawEscapeHatch(t *testing.T) {
	card := compileTo(t, `
elements:
  - raw:
      tag: person
      open_id: ou_xxx
`)
	e := mustElements(t, card)[0].(map[string]any)
	if e["tag"] != "person" || e["open_id"] != "ou_xxx" {
		t.Errorf("raw passthrough lost: %#v", e)
	}
}

func TestCompile_RawWithoutTagFails(t *testing.T) {
	_, err := Compile([]byte(`
elements:
  - raw: { open_id: ou_x }
`))
	if err == nil || !strings.Contains(err.Error(), "must include a 'tag' field") {
		t.Errorf("error = %v, want missing tag", err)
	}
}

func TestCompile_UnknownKind(t *testing.T) {
	_, err := Compile([]byte(`
elements:
  - frobnicate: "what"
`))
	if err == nil || !strings.Contains(err.Error(), "unknown element kind") {
		t.Errorf("error = %v, want unknown element kind", err)
	}
}

func TestCompile_EmptyElements(t *testing.T) {
	_, err := Compile([]byte(`elements: []`))
	if err == nil || !strings.Contains(err.Error(), "elements is empty") {
		t.Errorf("error = %v, want empty elements", err)
	}
}

func TestCompile_DisableWideScreen(t *testing.T) {
	card := compileTo(t, `
wide: false
update_multi: false
elements:
  - md: "x"
`)
	cfg := card["config"].(map[string]any)
	if cfg["wide_screen_mode"] != false {
		t.Errorf("wide_screen_mode = %v", cfg["wide_screen_mode"])
	}
	if cfg["update_multi"] != false {
		t.Errorf("update_multi = %v", cfg["update_multi"])
	}
}

func TestCompile_NestedPanelInPanel(t *testing.T) {
	card := compileTo(t, `
elements:
  - panel:
      title: "Outer"
      border: blue
      expanded: true
      elements:
        - md: "outer body"
        - panel:
            title: "Inner"
            border: red
            elements:
              - md: "inner body"
              - actions:
                  - { text: Ack, value: ack }
`)
	outer := mustElements(t, card)[0].(map[string]any)
	if outer["tag"] != "collapsible_panel" {
		t.Fatalf("outer = %#v", outer)
	}
	outerChildren := outer["elements"].([]any)
	if len(outerChildren) != 2 {
		t.Fatalf("outer children = %d, want 2", len(outerChildren))
	}
	inner := outerChildren[1].(map[string]any)
	if inner["tag"] != "collapsible_panel" {
		t.Fatalf("inner = %#v", inner)
	}
	if inner["border"].(map[string]any)["color"] != "red" {
		t.Errorf("inner border = %v, want red", inner["border"])
	}
	innerChildren := inner["elements"].([]any)
	if len(innerChildren) != 2 {
		t.Errorf("inner children = %d, want 2 (md + actions)", len(innerChildren))
	}
	if innerChildren[1].(map[string]any)["tag"] != "action" {
		t.Errorf("inner actions wrapper missing: %#v", innerChildren[1])
	}
}

func TestCompile_ExplicitlySetWideAndUpdateMulti(t *testing.T) {
	card := compileTo(t, `
wide: false
update_multi: false
elements:
  - md: "x"
`)
	cfg := card["config"].(map[string]any)
	if cfg["wide_screen_mode"] != false {
		t.Errorf("explicit wide=false should appear: %#v", cfg)
	}
	if cfg["update_multi"] != false {
		t.Errorf("explicit update_multi=false should appear: %#v", cfg)
	}
}

func TestCompile_HeaderlessIsValid(t *testing.T) {
	// Many schema 2.0 cards in the wild have no header — first
	// element is a markdown that serves as the title. Make sure
	// we don't accidentally require a header.
	card := compileTo(t, `
summary: "no header here"
elements:
  - md: "📋 **Self-titled card**"
  - hr
  - md: "body"
`)
	if _, has := card["header"]; has {
		t.Errorf("header should be absent when not declared, got %#v", card["header"])
	}
	if mustElements(t, card)[0].(map[string]any)["tag"] != "markdown" {
		t.Errorf("first element should be markdown")
	}
}

func TestCompile_MdSize_RejectsUnverified(t *testing.T) {
	// We only whitelist sizes both bridges actually emit. "heading"
	// and "title" appear in some Feishu docs but aren't used by the
	// reference bridges, so we reject them rather than silently emit
	// an unverified field.
	_, err := Compile([]byte(`
elements:
  - md:
      content: "x"
      size: heading
`))
	if err == nil || !strings.Contains(err.Error(), "invalid md.size") {
		t.Errorf("error = %v, want invalid md.size", err)
	}
}

func TestCompile_SchemaIsAlwaysV2(t *testing.T) {
	// Even a minimal card must declare schema 2.0 — without it
	// collapsible_panel falls back to v1 rendering and is shown
	// as an unknown element on Feishu's side.
	card := compileTo(t, `elements: [{ md: "x" }]`)
	if card["schema"] != "2.0" {
		t.Errorf("schema = %v, want 2.0", card["schema"])
	}
	// body wrapper is mandatory in v2
	if _, has := card["body"]; !has {
		t.Errorf("body wrapper missing")
	}
	if _, has := card["elements"]; has {
		t.Errorf("top-level elements is v1; should be under body in v2")
	}
}

// ─────────────────────────────────────────────────────────────────
// Tier-1 additions (openclaw-lark parity): markdown style optimizer,
// i18n_content, text_color, div+icon element, icon color/size.
// ─────────────────────────────────────────────────────────────────

func TestCompile_MarkdownStyle_DemotesH1ToH4(t *testing.T) {
	// Feishu renders raw H1/H2/H3 as gigantic text that breaks card
	// layout — the openclaw-lark builder explicitly demotes them. We
	// auto-apply the same demotion on every md element so a user
	// who pastes "# Title" in YAML doesn't sabotage their card.
	card := compileTo(t, `
elements:
  - md: |
      # H1 title
      ## H2 section
      ### H3 subsection
      Some body text.
`)
	content := mustElements(t, card)[0].(map[string]any)["content"].(string)
	if strings.Contains(content, "\n# ") || strings.HasPrefix(content, "# ") {
		t.Errorf("H1 not demoted, content still starts with '# ':\n%s", content)
	}
	if !strings.Contains(content, "#### H1 title") {
		t.Errorf("H1 should demote to H4 (####), got:\n%s", content)
	}
	if !strings.Contains(content, "##### H2 section") {
		t.Errorf("H2 should demote to H5 (#####), got:\n%s", content)
	}
}

func TestCompile_MarkdownStyle_NoStyleOptOut(t *testing.T) {
	// Power users (e.g. emitting pre-styled markdown from a higher
	// layer) can opt out of the auto-optimizer.
	card := compileTo(t, `
elements:
  - md:
      content: "# Big title"
      no_style: true
`)
	content := mustElements(t, card)[0].(map[string]any)["content"].(string)
	if !strings.HasPrefix(content, "# Big title") {
		t.Errorf("no_style=true should preserve H1, got:\n%s", content)
	}
}

func TestCompile_MarkdownStyle_StripsInvalidImageRefs(t *testing.T) {
	// Without this, ![alt](http://x) triggers CardKit error 200570 at
	// the Feishu side, killing the whole card. Mirror openclaw's
	// stripInvalidImageKeys safety net.
	card := compileTo(t, `
elements:
  - md: |
      Valid: ![ok](img_validkey)
      URL leaks: ![nope](https://example.com/x.png)
      Local: ![local](./x.png)
`)
	content := mustElements(t, card)[0].(map[string]any)["content"].(string)
	if !strings.Contains(content, "![ok](img_validkey)") {
		t.Errorf("valid img_xxx ref should be preserved, got:\n%s", content)
	}
	if strings.Contains(content, "example.com") {
		t.Errorf("URL image ref should be stripped:\n%s", content)
	}
	if strings.Contains(content, "./x.png") {
		t.Errorf("local-path image ref should be stripped:\n%s", content)
	}
}

func TestCompile_MarkdownStyle_PreservesCodeBlockContent(t *testing.T) {
	// Code blocks must round-trip exactly — H1 demotion inside a code
	// block would corrupt the user's example code.
	src := `
elements:
  - md: |
      Body text.

      ` + "```python" + `
      # not a real header — this is python comment
      def foo():
          return 1
      ` + "```" + `

      More body.
`
	card := compileTo(t, src)
	content := mustElements(t, card)[0].(map[string]any)["content"].(string)
	if !strings.Contains(content, "# not a real header") {
		t.Errorf("python comment inside code block was demoted:\n%s", content)
	}
	if !strings.Contains(content, "def foo():") {
		t.Errorf("code block body lost:\n%s", content)
	}
}

func TestCompile_I18nContent_HeaderAndSummary(t *testing.T) {
	card := compileTo(t, `
header:
  title: "Hello"
  title_i18n: { zh_cn: "你好", en_us: "Hello" }
  subtitle: "World"
  subtitle_i18n: { zh_cn: "世界" }
  template: blue
summary: "Quick brief"
summary_i18n: { zh_cn: "简报", en_us: "Quick brief" }
locales: [zh_cn, en_us]
elements:
  - md: "body"
`)
	title := card["header"].(map[string]any)["title"].(map[string]any)
	if title["i18n_content"].(map[string]any)["zh_cn"] != "你好" {
		t.Errorf("header title i18n missing: %#v", title)
	}
	sub := card["header"].(map[string]any)["subtitle"].(map[string]any)
	if sub["i18n_content"].(map[string]any)["zh_cn"] != "世界" {
		t.Errorf("header subtitle i18n missing: %#v", sub)
	}
	cfg := card["config"].(map[string]any)
	summary := cfg["summary"].(map[string]any)
	if summary["i18n_content"].(map[string]any)["zh_cn"] != "简报" {
		t.Errorf("summary i18n missing: %#v", summary)
	}
	locales := cfg["locales"].([]any)
	if len(locales) != 2 || locales[0] != "zh_cn" {
		t.Errorf("locales = %#v, want [zh_cn, en_us]", locales)
	}
}

func TestCompile_I18nContent_OnMarkdownElement(t *testing.T) {
	card := compileTo(t, `
elements:
  - md:
      content: "Hello world"
      i18n:
        zh_cn: "你好世界"
        en_us: "Hello world"
`)
	e := mustElements(t, card)[0].(map[string]any)
	i18n, ok := e["i18n_content"].(map[string]any)
	if !ok {
		t.Fatalf("md i18n_content missing: %#v", e)
	}
	if i18n["zh_cn"] != "你好世界" {
		t.Errorf("md i18n zh_cn = %v", i18n["zh_cn"])
	}
}

func TestCompile_TextColor_OnHeader(t *testing.T) {
	card := compileTo(t, `
header:
  title: "Muted"
  title_color: grey
  subtitle: "Caption"
  subtitle_color: turquoise
elements:
  - md: x
`)
	title := card["header"].(map[string]any)["title"].(map[string]any)
	if title["text_color"] != "grey" {
		t.Errorf("title.text_color = %v, want grey", title["text_color"])
	}
	sub := card["header"].(map[string]any)["subtitle"].(map[string]any)
	if sub["text_color"] != "turquoise" {
		t.Errorf("subtitle.text_color = %v, want turquoise", sub["text_color"])
	}
}

func TestCompile_TextColor_InvalidRejected(t *testing.T) {
	_, err := Compile([]byte(`
header:
  title: x
  title_color: rainbow
elements: [{ md: x }]
`))
	if err == nil || !strings.Contains(err.Error(), "invalid header.title_color") {
		t.Errorf("error = %v, want invalid title_color", err)
	}
}

func TestCompile_PanelTitleColor_OnlyOnPlainText(t *testing.T) {
	// text_color is silently dropped by Feishu on markdown titles. We
	// honour it only when title_tag is plain_text.
	card := compileTo(t, `
elements:
  - panel:
      title: "Plain title"
      title_tag: plain_text
      title_color: red
      border: red
      elements: [{ md: x }]
`)
	header := mustElements(t, card)[0].(map[string]any)["header"].(map[string]any)
	title := header["title"].(map[string]any)
	if title["tag"] != "plain_text" {
		t.Errorf("title.tag = %v, want plain_text", title["tag"])
	}
	if title["text_color"] != "red" {
		t.Errorf("panel title text_color = %v, want red", title["text_color"])
	}

	// And when title_tag defaults to markdown, text_color is dropped.
	card2 := compileTo(t, `
elements:
  - panel:
      title: "Md title"
      title_color: red
      elements: [{ md: x }]
`)
	header2 := mustElements(t, card2)[0].(map[string]any)["header"].(map[string]any)
	title2 := header2["title"].(map[string]any)
	if title2["tag"] != "markdown" {
		t.Errorf("default title.tag = %v, want markdown", title2["tag"])
	}
	if _, has := title2["text_color"]; has {
		t.Errorf("text_color should be dropped on markdown title: %#v", title2)
	}
}

func TestCompile_DivElement_WithIconAndColor(t *testing.T) {
	card := compileTo(t, `
elements:
  - div:
      text: "Running…"
      text_color: turquoise
      text_size: notation
      icon: standard:play-fill_filled
      icon_color: turquoise
      icon_size: "16px 16px"
`)
	e := mustElements(t, card)[0].(map[string]any)
	if e["tag"] != "div" {
		t.Fatalf("div tag = %v", e["tag"])
	}
	text := e["text"].(map[string]any)
	if text["tag"] != "plain_text" || text["content"] != "Running…" {
		t.Errorf("div text = %#v", text)
	}
	if text["text_color"] != "turquoise" {
		t.Errorf("div text_color = %v", text["text_color"])
	}
	if text["text_size"] != "notation" {
		t.Errorf("div text_size = %v", text["text_size"])
	}
	icon := e["icon"].(map[string]any)
	if icon["tag"] != "standard_icon" || icon["token"] != "play-fill_filled" {
		t.Errorf("div icon = %#v", icon)
	}
	if icon["color"] != "turquoise" || icon["size"] != "16px 16px" {
		t.Errorf("div icon color/size missing: %#v", icon)
	}
}

func TestCompile_DivElement_LarkMdAndI18n(t *testing.T) {
	card := compileTo(t, `
elements:
  - div:
      text: "**bold lark_md**"
      text_tag: lark_md
      text_i18n: { zh_cn: "**粗体**", en_us: "**bold**" }
`)
	text := mustElements(t, card)[0].(map[string]any)["text"].(map[string]any)
	if text["tag"] != "lark_md" {
		t.Errorf("text.tag = %v, want lark_md", text["tag"])
	}
	if text["i18n_content"].(map[string]any)["zh_cn"] != "**粗体**" {
		t.Errorf("div text i18n missing: %#v", text)
	}
}

func TestCompile_DivElement_InvalidTextTag(t *testing.T) {
	_, err := Compile([]byte(`
elements:
  - div: { text: "x", text_tag: markdown }
`))
	if err == nil || !strings.Contains(err.Error(), "invalid div.text_tag") {
		t.Errorf("error = %v, want invalid div.text_tag (only plain_text and lark_md allowed)", err)
	}
}

func TestCompile_HeaderIconColorAndSize(t *testing.T) {
	card := compileTo(t, `
header:
  title: "X"
  icon: standard:notice-bell
  icon_color: red
  icon_size: "20px 20px"
elements:
  - md: x
`)
	icon := card["header"].(map[string]any)["icon"].(map[string]any)
	if icon["tag"] != "standard_icon" || icon["token"] != "notice-bell" {
		t.Errorf("icon = %#v", icon)
	}
	if icon["color"] != "red" {
		t.Errorf("icon.color = %v", icon["color"])
	}
	if icon["size"] != "20px 20px" {
		t.Errorf("icon.size = %v", icon["size"])
	}
}

func TestCompile_CustomIconWithImageKey(t *testing.T) {
	card := compileTo(t, `
header:
  title: X
  icon: "custom:img_abc123"
elements: [{ md: x }]
`)
	icon := card["header"].(map[string]any)["icon"].(map[string]any)
	if icon["tag"] != "custom_icon" || icon["img_key"] != "img_abc123" {
		t.Errorf("custom icon = %#v", icon)
	}
}

func TestCompile_MarkdownAlignAndMargin(t *testing.T) {
	card := compileTo(t, `
elements:
  - md:
      content: "centered"
      align: center
      margin: "0px 0px 8px 0px"
`)
	e := mustElements(t, card)[0].(map[string]any)
	if e["text_align"] != "center" {
		t.Errorf("text_align = %v, want center", e["text_align"])
	}
	if e["margin"] != "0px 0px 8px 0px" {
		t.Errorf("margin = %v", e["margin"])
	}
}

func TestCompile_MarkdownAlignInvalid(t *testing.T) {
	_, err := Compile([]byte(`
elements:
  - md: { content: x, align: middle }
`))
	if err == nil || !strings.Contains(err.Error(), "invalid md.align") {
		t.Errorf("error = %v, want invalid md.align", err)
	}
}

// ─────────────────────────────────────────────────────────────────
// Round-3 polish: uniform markdown optimisation + auto-locales.
// ─────────────────────────────────────────────────────────────────

func TestCompile_PanelTitleMarkdownIsOptimised(t *testing.T) {
	// Panel titles using the default title_tag (markdown) should
	// go through the same H1→H4 demotion as md elements.
	card := compileTo(t, `
elements:
  - panel:
      title: "# Huge title would break layout"
      elements:
        - md: "body"
`)
	panelTitle := mustElements(t, card)[0].(map[string]any)["header"].(map[string]any)["title"].(map[string]any)
	content := panelTitle["content"].(string)
	if !strings.Contains(content, "#### Huge title") {
		t.Errorf("panel title H1 not demoted, got: %q", content)
	}
}

func TestCompile_PanelTitlePlainTextNotOptimised(t *testing.T) {
	// plain_text titles must NOT go through the markdown optimiser —
	// plain_text doesn't parse markdown syntax, so demoting H1 would
	// just insert literal '####' characters into the rendered text.
	card := compileTo(t, `
elements:
  - panel:
      title: "# Literal hash should stay"
      title_tag: plain_text
      elements: [{ md: x }]
`)
	title := mustElements(t, card)[0].(map[string]any)["header"].(map[string]any)["title"].(map[string]any)
	if title["content"] != "# Literal hash should stay" {
		t.Errorf("plain_text title was incorrectly optimised: %v", title["content"])
	}
}

func TestCompile_DivLarkMdIsOptimised(t *testing.T) {
	card := compileTo(t, `
elements:
  - div:
      text: "# Big lark_md heading"
      text_tag: lark_md
`)
	text := mustElements(t, card)[0].(map[string]any)["text"].(map[string]any)
	content := text["content"].(string)
	if !strings.Contains(content, "#### Big lark_md heading") {
		t.Errorf("div lark_md not optimised: %v", content)
	}
}

func TestCompile_DivPlainTextNotOptimised(t *testing.T) {
	card := compileTo(t, `
elements:
  - div:
      text: "# Literal in plain_text div"
`)
	text := mustElements(t, card)[0].(map[string]any)["text"].(map[string]any)
	if text["content"] != "# Literal in plain_text div" {
		t.Errorf("plain_text div was incorrectly optimised: %v", text["content"])
	}
}

func TestCompile_MdCodeBlockGetsBrPadding(t *testing.T) {
	// Round-trip from the YAML level: md content with a fenced
	// code block should end up <br>-padded in the compiled output.
	card := compileTo(t, "elements:\n  - md: |\n      Body.\n\n      ```\n      foo\n      ```\n\n      More.\n")
	content := mustElements(t, card)[0].(map[string]any)["content"].(string)
	if !strings.Contains(content, "<br>") {
		t.Errorf("compiled md missing <br> padding around code block:\n%s", content)
	}
}

func TestCompile_AutoLocalesFromMdI18n(t *testing.T) {
	// When the spec uses i18n_content anywhere but doesn't declare
	// `locales:`, the compiler auto-derives them from observed keys.
	card := compileTo(t, `
elements:
  - md:
      content: "hello"
      i18n: { zh_cn: "你好", en_us: "hello" }
`)
	cfg := card["config"].(map[string]any)
	locales, ok := cfg["locales"].([]any)
	if !ok {
		t.Fatalf("locales not auto-derived: %#v", cfg)
	}
	if len(locales) != 2 || locales[0] != "en_us" || locales[1] != "zh_cn" {
		// Alphabetical order is the contract.
		t.Errorf("locales = %#v, want [en_us, zh_cn] (sorted)", locales)
	}
}

func TestCompile_AutoLocalesUnionAcrossElements(t *testing.T) {
	// Auto-derivation walks the whole tree (header, summary, md,
	// panel.title, div.text) and unions every locale key it sees.
	card := compileTo(t, `
header:
  title: "T"
  title_i18n: { en_us: "T" }
summary: "S"
summary_i18n: { ja_jp: "S" }
elements:
  - panel:
      title: "P"
      title_i18n: { zh_cn: "P" }
      elements:
        - div:
            text: "D"
            text_tag: lark_md
            text_i18n: { en_us: "D", ko_kr: "D" }
`)
	cfg := card["config"].(map[string]any)
	locales := cfg["locales"].([]any)
	got := make([]string, len(locales))
	for i, v := range locales {
		got[i] = v.(string)
	}
	wantSet := map[string]bool{"en_us": true, "ja_jp": true, "zh_cn": true, "ko_kr": true}
	if len(got) != len(wantSet) {
		t.Fatalf("auto-locales = %v, want union of {en_us, ja_jp, zh_cn, ko_kr}", got)
	}
	for _, locale := range got {
		if !wantSet[locale] {
			t.Errorf("unexpected locale %q in auto-derived list", locale)
		}
	}
	// Verify sorted alphabetical
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Errorf("auto-locales not sorted: %v", got)
			break
		}
	}
}

func TestCompile_ExplicitLocalesOverridesAuto(t *testing.T) {
	// When `locales:` is set explicitly, the auto-derivation is
	// skipped — users who want to ship a stricter list can.
	card := compileTo(t, `
locales: [en_us]
elements:
  - md:
      content: "hello"
      i18n: { zh_cn: "你好", en_us: "hello" }
`)
	cfg := card["config"].(map[string]any)
	locales := cfg["locales"].([]any)
	if len(locales) != 1 || locales[0] != "en_us" {
		t.Errorf("explicit locales should win: got %v, want [en_us]", locales)
	}
}

func TestCompile_NoLocalesWhenNoI18nUsed(t *testing.T) {
	card := compileTo(t, `elements: [{ md: hello }]`)
	cfg := card["config"].(map[string]any)
	if _, has := cfg["locales"]; has {
		t.Errorf("locales should be absent when no i18n is used: %#v", cfg)
	}
}

// ─────────────────────────────────────────────────────────────────
// Coverage-driven additions: error paths in UnmarshalYAML that were
// previously unreached, plus a Compile() idempotency invariant.
// ─────────────────────────────────────────────────────────────────

func TestCompile_ElementNotAMapping(t *testing.T) {
	// `- 42` is a scalar that isn't the "hr" shorthand. Should error
	// with a precise message instead of producing garbage downstream.
	_, err := Compile([]byte(`
elements:
  - 42
`))
	if err == nil || !strings.Contains(err.Error(), "element must be a mapping or the scalar 'hr'") {
		t.Errorf("error = %v, want non-mapping element message", err)
	}
}

func TestCompile_EmptyElementMapping(t *testing.T) {
	// `- {}` is technically a mapping but has no discriminator key.
	// Should error rather than silently producing an unknown element.
	_, err := Compile([]byte(`
elements:
  - {}
`))
	if err == nil || !strings.Contains(err.Error(), "empty element") {
		t.Errorf("error = %v, want empty-element message", err)
	}
}

func TestCompile_InvalidDivIconColor_IncludesAllowedList(t *testing.T) {
	// Regression guard: every "invalid X" error should include the
	// (allowed: ...) hint. Previously div.icon_color was the one
	// outlier without the hint — a UX inconsistency that made the
	// error noticeably less helpful than its siblings.
	_, err := Compile([]byte(`
elements:
  - div:
      text: x
      icon: standard:bell
      icon_color: rainbow
`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "(allowed:") {
		t.Errorf("div.icon_color error missing (allowed: ...) hint: %v", err)
	}
}

func TestCompile_InvalidHeaderSubtitleColor_IncludesAllowedList(t *testing.T) {
	_, err := Compile([]byte(`
header:
  title: T
  subtitle: S
  subtitle_color: rainbow
elements: [{ md: x }]
`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "(allowed:") {
		t.Errorf("header.subtitle_color error missing (allowed: ...) hint: %v", err)
	}
}

func TestCompile_InvalidHeaderIconColor_IncludesAllowedList(t *testing.T) {
	_, err := Compile([]byte(`
header:
  title: T
  icon: standard:notice-bell
  icon_color: rainbow
elements: [{ md: x }]
`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "(allowed:") {
		t.Errorf("header.icon_color error missing (allowed: ...) hint: %v", err)
	}
}

// TestCompile_IsIdempotent guards against hidden mutable state in
// the compile path (regex caches, sorting, map iteration order).
// Two invocations on the same spec must produce byte-identical
// output — if this ever fails, fixtures and diff-tests downstream
// stop being meaningful.
func TestCompile_IsIdempotent(t *testing.T) {
	specs := []string{
		`elements: [{ md: x }]`,
		`
header:
  title: "T"
  template: blue
summary: "S"
elements:
  - md:
      content: "# H1 demotes"
      i18n: { zh_cn: "标题", en_us: "Title", ja_jp: "タイトル" }
  - div:
      text: "Build"
      icon: standard:check-circle_filled
      icon_color: green
  - panel:
      title: "P"
      title_i18n: { zh_cn: "面板" }
      elements:
        - md: "body"
`,
	}
	for _, src := range specs {
		first, err := Compile([]byte(src))
		if err != nil {
			t.Fatalf("Compile #1 error = %v", err)
		}
		second, err := Compile([]byte(src))
		if err != nil {
			t.Fatalf("Compile #2 error = %v", err)
		}
		if first != second {
			t.Errorf("Compile is not idempotent for spec:\n%s\n\nfirst != second", src)
		}
	}
}

// TestCompile_ScreenshotReport reproduces the actual report from the
// user's screenshot: 3 chat groups about Lark AI Certificate, each
// rendered as a panel with chat_id, summary, action buttons. This is
// the canonical "real" example the skill docs reference.
func TestCompile_ScreenshotReport(t *testing.T) {
	src := `
header:
  title: "📊 Lark AI Certificate — 3 nhóm chat"
  template: blue
summary: "3 chat về AI Cert — 1 nội bộ, 2 external"
elements:
  - md: "**Nhóm chính (rõ ràng nhất)**"
  - panel:
      title: "[IMPORTANT] Chứng nhận AI Lark"
      border: red
      expanded: true
      elements:
        - md: |
            - chat_id: ` + "`oc_24689d68…`" + `
            - Tạo 2026-05-27 10:55, giao @Sơn + @Sang
            - Đối tượng: Channel Sales / Presales / CSM partner
        - actions:
            - text: "Mở chat"
              value: "open:oc_24689d68"
              style: primary
            - text: "Mở doc"
              url: "https://bytedance.my.larkoffice.com/docx/CDiud369ZoR8S7xwPXRmzLQmyDd"
          layout: equal
  - hr
  - panel:
      title: "Chat partner-comms (external)"
      border: grey
      elements:
        - md: "rep từ Lark (chị Jenny) — 2026-05-27 10:24"
        - note: "Cùng kênh Partner Newsletter từ 2025-02"
`
	card := compileTo(t, src)
	elems := mustElements(t, card)
	if len(elems) != 4 {
		t.Fatalf("screenshot report element count = %d, want 4", len(elems))
	}
	// First panel must be red + expanded
	p1 := elems[1].(map[string]any)
	if p1["tag"] != "collapsible_panel" || p1["expanded"] != true {
		t.Errorf("panel[1] = %#v", p1)
	}
	if p1["border"].(map[string]any)["color"] != "red" {
		t.Errorf("panel[1] border not red")
	}
	// Panel actions: 2 buttons in equal layout = column_set with bisect flex
	p1Children := p1["elements"].([]any)
	actionsEl := p1Children[1].(map[string]any)
	if actionsEl["tag"] != "column_set" || actionsEl["flex_mode"] != "bisect" {
		t.Errorf("panel actions not equal layout: %#v", actionsEl)
	}
}
